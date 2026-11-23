package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"sync"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

const (
	StateUploading  = "uploading"
	StateAssembling = "assembling"
	StateStored     = "stored"

	assembleHashMismatch = "PACK_FILE_HASH_MISMATCH"
	assemblePartLost     = "PACK_PART_LOST"
	assembleStorageError = "STORAGE_ERROR"
)

var errAssembleHash = errors.New("assembled file does not match its sha256")

type partLostError struct{ part int }

func (e *partLostError) Error() string { return fmt.Sprintf("part %d is missing from storage", e.part) }

type assembler struct {
	mu     sync.Mutex
	queue  []string
	queued map[string]bool
	active int
	wg     sync.WaitGroup
}

func newAssembler() *assembler { return &assembler{queued: map[string]bool{}} }

func (s *service) kickAssembly(shas ...string) {
	if len(shas) == 0 {
		return
	}
	limit := s.settings(s.bg).Packs.AssembleWorkers
	a := s.asm
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, sha := range shas {
		if a.queued[sha] {
			continue
		}
		a.queued[sha] = true
		a.queue = append(a.queue, sha)
	}
	s.pumpLocked(limit)
}

func (s *service) pumpLocked(limit int) {
	a := s.asm
	for a.active < limit && len(a.queue) > 0 {
		sha := a.queue[0]
		a.queue = a.queue[1:]
		a.active++
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			if err := s.assembleBlob(s.bg, sha); err != nil && s.logger != nil {
				s.logger.Warn("Content blob %s was not assembled: %v", sha, err)
			}
			limit := s.settings(s.bg).Packs.AssembleWorkers
			a.mu.Lock()
			a.active--
			delete(a.queued, sha)
			s.pumpLocked(limit)
			a.mu.Unlock()
		}()
	}
}

func (s *service) waitAssembly() { s.asm.wg.Wait() }

func (s *service) resumeAssemblies(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT sha256 FROM content_blob WHERE NOT stored AND cardinality(received)=parts
		AND (assemble_until IS NULL OR assemble_until < now()) ORDER BY created_at LIMIT 100`)
	if err != nil {
		return err
	}
	shas := []string{}
	for rows.Next() {
		var sha string
		if err := rows.Scan(&sha); err != nil {
			rows.Close()
			return err
		}
		shas = append(shas, sha)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	s.kickAssembly(shas...)
	return nil
}

func (s *service) assembleLoop(ctx context.Context) {
	if err := s.resumeAssemblies(ctx); err != nil && s.logger != nil {
		s.logger.Warn("Content assembly resume failed: %v", err)
	}
	for {
		wait := time.Duration(s.settings(ctx).Packs.AssembleSweepSec) * time.Second
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
			if err := s.resumeAssemblies(ctx); err != nil && s.logger != nil {
				s.logger.Warn("Content assembly sweep failed: %v", err)
			}
		}
	}
}

type partsReader struct {
	ctx   context.Context
	store objectstore.Store
	sha   string
	parts int
	next  int
	cur   io.ReadCloser
}

func (r *partsReader) Read(p []byte) (int, error) {
	for {
		if r.cur == nil {
			if r.next >= r.parts {
				return 0, io.EOF
			}
			obj, err := r.store.Get(r.ctx, partKey(r.sha, r.next))
			if errors.Is(err, objectstore.ErrNotFound) {
				return 0, &partLostError{part: r.next}
			}
			if err != nil {
				return 0, err
			}
			r.cur = obj.Body
			r.next++
		}
		n, err := r.cur.Read(p)
		if errors.Is(err, io.EOF) {
			r.cur.Close()
			r.cur = nil
			if n > 0 {
				return n, nil
			}
			continue
		}
		return n, err
	}
}

func (r *partsReader) Close() {
	if r.cur != nil {
		r.cur.Close()
		r.cur = nil
	}
}

type verifyReader struct {
	r    io.Reader
	h    hash.Hash
	n    int64
	size int64
	sha  string
}

func (v *verifyReader) Read(p []byte) (int, error) {
	n, err := v.r.Read(p)
	v.h.Write(p[:n])
	v.n += int64(n)
	if v.n > v.size {
		return n, errAssembleHash
	}
	if errors.Is(err, io.EOF) && (v.n != v.size || hex.EncodeToString(v.h.Sum(nil)) != v.sha) {
		return n, errAssembleHash
	}
	return n, err
}

func (s *service) assembleBlob(ctx context.Context, sha string) error {
	cfg := s.settings(ctx).Packs
	var size int64
	var parts int
	var ctype string
	err := s.db.QueryRowContext(ctx, `UPDATE content_blob SET assemble_holder=$2, assemble_until=now() + make_interval(secs => $3), assemble_started_at=now(),
			assemble_attempts=assemble_attempts+1, assemble_error=''
		WHERE sha256=$1 AND NOT stored AND cardinality(received)=parts AND (assemble_until IS NULL OR assemble_until < now())
		RETURNING size, parts, content_type`, sha, s.holder, cfg.AssembleLeaseSec).Scan(&size, &parts, &ctype)
	if isNoRows(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if s.assembleHook != nil {
		s.assembleHook(sha)
	}
	actx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	go s.renewLease(actx, cancel, sha, cfg.AssembleLeaseSec)
	src := &partsReader{ctx: actx, store: s.store, sha: sha, parts: parts}
	err = objectstore.PutStream(actx, s.store, blobKey(sha), &verifyReader{r: src, h: sha256.New(), size: size, sha: sha}, ctype)
	src.Close()
	if cause := context.Cause(actx); err != nil && cause != nil && !errors.Is(cause, context.Canceled) {
		err = cause
	}
	dbctx, dbcancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer dbcancel()
	var lost *partLostError
	switch {
	case err == nil:
		if _, err := s.db.ExecContext(dbctx, `UPDATE content_blob SET stored=true, stored_at=now(), assemble_holder='', assemble_until=NULL, assemble_error=''
			WHERE sha256=$1 AND NOT stored`, sha); err != nil {
			return err
		}
		s.deleteParts(dbctx, sha, parts)
		return nil
	case errors.Is(err, errAssembleHash):
		if !s.holdsLease(dbctx, sha) {
			return errLeaseLost
		}
		s.deleteParts(dbctx, sha, parts)
		res, uerr := s.db.ExecContext(dbctx, `UPDATE content_blob SET received='{}', assemble_holder='', assemble_until=NULL, assemble_error=$3
			WHERE sha256=$1 AND assemble_holder=$2 AND NOT stored`, sha, s.holder, assembleHashMismatch)
		if uerr != nil {
			return uerr
		}
		if n, _ := res.RowsAffected(); n > 0 {
			s.rejectEvents(dbctx, sha, assembleHashMismatch, -1)
		}
		return err
	case errors.As(err, &lost):
		res, uerr := s.db.ExecContext(dbctx, `UPDATE content_blob SET received=array_remove(received, $3::int), assemble_holder='', assemble_until=NULL, assemble_error=$4
			WHERE sha256=$1 AND assemble_holder=$2 AND NOT stored`, sha, s.holder, lost.part, assemblePartLost)
		if uerr != nil {
			return uerr
		}
		if n, _ := res.RowsAffected(); n > 0 {
			s.rejectEvents(dbctx, sha, assemblePartLost, lost.part)
		}
		return err
	default:
		if _, uerr := s.db.ExecContext(dbctx, `UPDATE content_blob SET assemble_holder='', assemble_error=$3,
				assemble_until=now() + make_interval(secs => LEAST(600, 15 * assemble_attempts))
			WHERE sha256=$1 AND assemble_holder=$2 AND NOT stored`, sha, s.holder, assembleStorageError); uerr != nil {
			return uerr
		}
		return err
	}
}

var errLeaseLost = errors.New("assembly lease was taken by another node")

func (s *service) holdsLease(ctx context.Context, sha string) bool {
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM content_blob WHERE sha256=$1 AND assemble_holder=$2 AND NOT stored)`, sha, s.holder).Scan(&ok)
	return err == nil && ok
}

func (s *service) renewLease(ctx context.Context, cancel context.CancelCauseFunc, sha string, leaseSec int) {
	every := time.Duration(leaseSec) * time.Second / 3
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			res, err := s.db.ExecContext(ctx, `UPDATE content_blob SET assemble_until=now() + make_interval(secs => $3) WHERE sha256=$1 AND assemble_holder=$2 AND NOT stored`, sha, s.holder, leaseSec)
			if err != nil {
				continue
			}
			if n, _ := res.RowsAffected(); n == 0 {
				cancel(errLeaseLost)
				return
			}
		}
	}
}

func (s *service) rejectEvents(ctx context.Context, sha, code string, part int) {
	detail := map[string]any{"sha256": sha, "code": code}
	if part >= 0 {
		detail["part"] = part
	}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT p.id::text, p.platform, p.version, f.path FROM content_pack p JOIN content_pack_file f ON f.pack_id=p.id
		WHERE f.sha256=$1 AND p.status='uploading'`, sha)
	if err != nil {
		return
	}
	type ev struct{ id, platform, version, path string }
	list := []ev{}
	for rows.Next() {
		var e ev
		if rows.Scan(&e.id, &e.platform, &e.version, &e.path) == nil {
			list = append(list, e)
		}
	}
	rows.Close()
	for _, e := range list {
		detail["path"] = e.path
		raw, _ := json.Marshal(detail)
		_, _ = s.db.ExecContext(ctx, `INSERT INTO content_pack_event (pack_id, platform, version, action, actor, detail) VALUES ($1,$2,$3,'file_rejected','system',$4)`, e.id, e.platform, e.version, raw)
	}
}
