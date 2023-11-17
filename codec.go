package jce

import (
	"bytes"
	"errors"
	"io"
)

type Messager interface {
	io.ReaderFrom
	io.WriterTo
}

// Marshal to io.Writer
// tip: v need is a pointer
func MarshalTo(v any, w io.Writer) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.WriteTo(w)
	return
}

// Marshal
// tip: v need is a pointer
func Marshal(v any) (data []byte, err error) {
	b := bytes.NewBuffer(make([]byte, 0))
	if err = MarshalTo(v, b); err != nil {
		return
	}
	return b.Bytes(), nil
}

// Unmarshal from io.Reader
// tip: v need is a pointer
func UnmarshalFrom(r io.Reader, v any) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.ReadFrom(r)
	return
}

// Unmarshal
// tip: v need is a pointer
func Unmarshal(data []byte, v any) (err error) {
	return UnmarshalFrom(bytes.NewBuffer(data), v)
}
