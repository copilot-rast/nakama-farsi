package cluster

import (
	"encoding/binary"
	"errors"
)

var ErrShortFrame = errors.New("cluster: short frame")

type Enc struct {
	b []byte
}

func NewEnc(size int) *Enc {
	return &Enc{b: make([]byte, 0, size)}
}

func (e *Enc) U8(v byte) *Enc {
	e.b = append(e.b, v)
	return e
}

func (e *Enc) Bool(v bool) *Enc {
	if v {
		return e.U8(1)
	}
	return e.U8(0)
}

func (e *Enc) Uvarint(v uint64) *Enc {
	e.b = binary.AppendUvarint(e.b, v)
	return e
}

func (e *Enc) Varint(v int64) *Enc {
	e.b = binary.AppendVarint(e.b, v)
	return e
}

func (e *Enc) Bytes(v []byte) *Enc {
	e.b = binary.AppendUvarint(e.b, uint64(len(v)))
	e.b = append(e.b, v...)
	return e
}

func (e *Enc) Str(v string) *Enc {
	e.b = binary.AppendUvarint(e.b, uint64(len(v)))
	e.b = append(e.b, v...)
	return e
}

func (e *Enc) Raw16(v [16]byte) *Enc {
	e.b = append(e.b, v[:]...)
	return e
}

func (e *Enc) Out() []byte {
	return e.b
}

type Dec struct {
	b   []byte
	err error
}

func NewDec(b []byte) *Dec {
	return &Dec{b: b}
}

func (d *Dec) Err() error {
	return d.err
}

func (d *Dec) Len() int {
	return len(d.b)
}

func (d *Dec) fail() {
	if d.err == nil {
		d.err = ErrShortFrame
	}
	d.b = nil
}

func (d *Dec) U8() byte {
	if len(d.b) < 1 {
		d.fail()
		return 0
	}
	v := d.b[0]
	d.b = d.b[1:]
	return v
}

func (d *Dec) Bool() bool {
	return d.U8() == 1
}

func (d *Dec) Uvarint() uint64 {
	v, n := binary.Uvarint(d.b)
	if n <= 0 {
		d.fail()
		return 0
	}
	d.b = d.b[n:]
	return v
}

func (d *Dec) Varint() int64 {
	v, n := binary.Varint(d.b)
	if n <= 0 {
		d.fail()
		return 0
	}
	d.b = d.b[n:]
	return v
}

func (d *Dec) Bytes() []byte {
	l := d.Uvarint()
	if d.err != nil || uint64(len(d.b)) < l {
		d.fail()
		return nil
	}
	v := d.b[:l:l]
	d.b = d.b[l:]
	return v
}

func (d *Dec) Str() string {
	return string(d.Bytes())
}

func (d *Dec) Raw16() [16]byte {
	var v [16]byte
	if len(d.b) < 16 {
		d.fail()
		return v
	}
	copy(v[:], d.b[:16])
	d.b = d.b[16:]
	return v
}

func AppendFrame(dst []byte, kind byte, payload []byte) []byte {
	dst = append(dst, kind)
	dst = binary.AppendUvarint(dst, uint64(len(payload)))
	return append(dst, payload...)
}

func SplitFrames(batch []byte, fn func(kind byte, payload []byte)) error {
	for len(batch) > 0 {
		kind := batch[0]
		l, n := binary.Uvarint(batch[1:])
		if n <= 0 || uint64(len(batch)-1-n) < l {
			return ErrShortFrame
		}
		start := 1 + n
		fn(kind, batch[start:start+int(l):start+int(l)])
		batch = batch[start+int(l):]
	}
	return nil
}
