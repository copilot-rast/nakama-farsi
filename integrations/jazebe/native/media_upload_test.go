package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateMediaUploadBegin(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	valid := mediaUploadBeginInput{PlanetID: id, Purpose: "evidence", FileName: `folder\photo.png`, ExpectedBytes: 25 * 1024 * 1024}
	got, err := validateMediaUploadBegin(valid)
	if err != nil || got.FileName != "photo.png" {
		t.Fatalf("valid evidence upload rejected: %#v, %v", got, err)
	}
	for _, change := range []func(*mediaUploadBeginInput){
		func(i *mediaUploadBeginInput) { i.Purpose = "avatar" },
		func(i *mediaUploadBeginInput) { i.ExpectedBytes++ },
		func(i *mediaUploadBeginInput) { i.ExpectedBytes = 0 },
		func(i *mediaUploadBeginInput) { i.FileName = "\n" },
		func(i *mediaUploadBeginInput) { i.PlanetID = "bad" },
	} {
		input := valid
		change(&input)
		if _, err := validateMediaUploadBegin(input); err == nil {
			t.Fatalf("invalid upload accepted: %#v", input)
		}
	}
	valid.Purpose, valid.ExpectedBytes = "mission_media", 50*1024*1024
	if _, err := validateMediaUploadBegin(valid); err != nil {
		t.Fatalf("mission media limit rejected: %v", err)
	}
	valid.Purpose, valid.ExpectedBytes = "avatar", 5*1024*1024
	if _, err := validateMediaUploadBegin(valid); err != nil {
		t.Fatalf("avatar limit rejected: %v", err)
	}
}

func TestMediaUploadRateConfiguration(t *testing.T) {
	t.Setenv("JAZEBE_MEDIA_UPLOAD_PER_MINUTE", "")
	if limit, err := mediaUploadPerMinute(); err != nil || limit != 20 {
		t.Fatalf("default rate: %d, %v", limit, err)
	}
	t.Setenv("JAZEBE_MEDIA_UPLOAD_PER_MINUTE", "3")
	if limit, err := mediaUploadPerMinute(); err != nil || limit != 3 {
		t.Fatalf("configured rate: %d, %v", limit, err)
	}
	for _, invalid := range []string{"0", "-1", "10001", "fast"} {
		t.Setenv("JAZEBE_MEDIA_UPLOAD_PER_MINUTE", invalid)
		if _, err := mediaUploadPerMinute(); err == nil {
			t.Fatalf("invalid rate %q accepted", invalid)
		}
	}
}

func TestMediaUploadDeleteRequiresSession(t *testing.T) {
	if _, err := mediaUploadDelete(context.Background(), nil, nil, nil, `{}`); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("anonymous media withdrawal accepted: %v", err)
	}
}

func TestSniffMedia(t *testing.T) {
	for _, tc := range []struct {
		data       []byte
		mime, kind string
	}{
		{[]byte("\x89PNG\r\n\x1a\n"), "image/png", "image"},
		{[]byte("\xff\xd8\xff\xe0"), "image/jpeg", "image"},
		{[]byte("GIF89a"), "image/gif", "image"},
		{[]byte("RIFF1234WEBP"), "image/webp", "image"},
		{[]byte("\x00\x00\x00\x18ftypisom"), "video/mp4", "video"},
		{[]byte("<script>alert(1)</script>"), "", ""},
		{[]byte("%PDF-1.7"), "", ""},
	} {
		mime, kind := sniffMedia(tc.data)
		if mime != tc.mime || kind != tc.kind {
			t.Fatalf("sniff %q: got %s %s", tc.data, mime, kind)
		}
	}
}

func TestMediaChunkRejectsMalformedRequestBeforeDatabase(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	for _, tc := range []struct {
		id, index string
		body      []byte
		status    int
	}{
		{"bad", "0", []byte("hello"), http.StatusNotFound},
		{id, "-1", []byte("hello"), http.StatusBadRequest},
		{id, "0", bytes.Repeat([]byte("x"), mediaChunkBytes+1), http.StatusRequestEntityTooLarge},
		{id, "0", nil, http.StatusRequestEntityTooLarge},
	} {
		r := httptest.NewRequest(http.MethodPost, "/v2/jazebe/media/"+tc.id+"/chunks", bytes.NewReader(tc.body))
		r.Header.Set("X-Chunk-Index", tc.index)
		r = mux.SetURLVars(r, map[string]string{"id": tc.id})
		w := httptest.NewRecorder()
		mediaChunkHTTP(nil)(w, r)
		if w.Code != tc.status || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("invalid chunk returned %d, expected %d", w.Code, tc.status)
		}
	}
}
