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

func Marshal(v any) (data []byte, err error) {
	m, ok := v.(Messager)
	if !ok {
		return nil, errors.New("not jce Messager type")
	}
	b := bytes.NewBuffer(make([]byte, 0))
	_, err = m.WriteTo(b)
	return b.Bytes(), err
}

func MarshalTo(v any, w io.Writer) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.WriteTo(w)
	return
}

func Unmarshal(data []byte, v any) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	b := bytes.NewBuffer(data)
	_, err = m.ReadFrom(b)
	return
}

func UnmarshalFrom(r io.Reader, v any) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.ReadFrom(r)
	return
}
