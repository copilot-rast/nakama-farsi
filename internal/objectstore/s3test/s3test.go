package s3test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const MinPartBytes = 5 << 20

type Server struct {
	mu        sync.Mutex
	Objects   map[string][]byte
	Uploads   map[string]map[int][]byte
	PartSizes []int
	Puts      int
	Aborted   int
	Completed int
	FailPart  int
	next      int
}

func New() *Server {
	return &Server{Objects: map[string][]byte{}, Uploads: map[string]map[int][]byte{}}
}

func (f *Server) Snapshot() (completed, puts, aborted, open int, partSizes []int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Completed, f.Puts, f.Aborted, len(f.Uploads), append([]int(nil), f.PartSizes...)
}

func (f *Server) Object(path string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok := f.Objects[path]
	return data, ok
}

func (f *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	sum := sha256.Sum256(body)
	if r.Header.Get("X-Amz-Content-Sha256") != hex.EncodeToString(sum[:]) || !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("<Error><Code>SignatureDoesNotMatch</Code></Error>"))
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	key := r.URL.Path
	q := r.URL.Query()
	switch {
	case r.Method == http.MethodPost && q.Has("uploads"):
		f.next++
		id := "up-" + strconv.Itoa(f.next)
		f.Uploads[id] = map[int][]byte{}
		_, _ = fmt.Fprintf(w, "<InitiateMultipartUploadResult><UploadId>%s</UploadId></InitiateMultipartUploadResult>", id)
	case r.Method == http.MethodPut && q.Get("uploadId") != "":
		parts, ok := f.Uploads[q.Get("uploadId")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		n, _ := strconv.Atoi(q.Get("partNumber"))
		if f.FailPart == n {
			f.FailPart = 0
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		parts[n] = body
		f.PartSizes = append(f.PartSizes, len(body))
		w.Header().Set("ETag", `"etag-`+strconv.Itoa(n)+`"`)
	case r.Method == http.MethodPost && q.Get("uploadId") != "":
		parts, ok := f.Uploads[q.Get("uploadId")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var in struct {
			Parts []struct {
				PartNumber int    `xml:"PartNumber"`
				ETag       string `xml:"ETag"`
			} `xml:"Part"`
		}
		if err := xml.Unmarshal(body, &in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var all bytes.Buffer
		for i, p := range in.Parts {
			if p.PartNumber != i+1 || p.ETag != `"etag-`+strconv.Itoa(p.PartNumber)+`"` {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if i < len(in.Parts)-1 && len(parts[p.PartNumber]) < MinPartBytes {
				_, _ = w.Write([]byte("<Error><Code>EntityTooSmall</Code></Error>"))
				return
			}
			all.Write(parts[p.PartNumber])
		}
		f.Objects[key] = all.Bytes()
		delete(f.Uploads, q.Get("uploadId"))
		f.Completed++
		_, _ = w.Write([]byte("<CompleteMultipartUploadResult><Key>x</Key></CompleteMultipartUploadResult>"))
	case r.Method == http.MethodDelete && q.Get("uploadId") != "":
		delete(f.Uploads, q.Get("uploadId"))
		f.Aborted++
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPut:
		f.Objects[key] = body
		f.Puts++
	case r.Method == http.MethodGet:
		data, ok := f.Objects[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write(data)
	case r.Method == http.MethodDelete:
		delete(f.Objects, key)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
