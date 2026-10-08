// Package codec is a minimal local stand-in for github.com/ugorji/go/codec.
//
// The real package is a huge generated codebase whose compilation was being
// OOM-killed in the build sandbox ("compile: signal: killed"). It is only
// pulled in by gin's optional MsgPack render/binding, which this application
// never uses. This stub provides the small API surface gin references so the
// build fits in memory. MsgPack (de)serialization falls back to JSON.
package codec

import (
	"encoding/json"
	"io"
)

// Handle is the marker interface for codec handles.
type Handle interface {
	isHandle()
}

// MsgpackHandle mirrors the fields of the upstream handle that callers may set.
type MsgpackHandle struct {
	RawToString bool
	WriteExt    bool
	Canonical   bool
}

func (*MsgpackHandle) isHandle() {}

// Encoder encodes values to a writer.
type Encoder struct{ w io.Writer }

// NewEncoder returns an Encoder writing to w.
func NewEncoder(w io.Writer, h Handle) *Encoder { return &Encoder{w: w} }

// Encode writes v.
func (e *Encoder) Encode(v interface{}) error { return json.NewEncoder(e.w).Encode(v) }

// Decoder decodes values from a reader.
type Decoder struct{ r io.Reader }

// NewDecoder returns a Decoder reading from r.
func NewDecoder(r io.Reader, h Handle) *Decoder { return &Decoder{r: r} }

// Decode reads into v.
func (d *Decoder) Decode(v interface{}) error { return json.NewDecoder(d.r).Decode(v) }
