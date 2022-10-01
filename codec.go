package jce

import (
	"errors"
	"io"
)

type Messager interface {
	io.ReaderFrom
	io.WriterTo
}

func Marshal(v any, w io.Writer) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.WriteTo(w)
	return
}

func Unmarshal(r io.Reader, v any) (err error) {
	m, ok := v.(Messager)
	if !ok {
		return errors.New("not jce Messager type")
	}
	_, err = m.ReadFrom(r)
	return
}
