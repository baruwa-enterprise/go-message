package mail

import (
	"container/list"
	"io"

	"github.com/emersion/go-message"
)

// NewReaderF creates a new tolerant mail reader.
func NewReaderF(e *message.Entity) *Reader {
	mr := e.MultipartReader()
	if mr == nil {
		// Artificially create a multipart entity
		// With this header, no error will be returned by message.NewMultipart
		var h message.Header
		h.Set("Content-Type", "multipart/mixed")
		me, _ := message.NewMultipartF(h, []*message.Entity{e})
		mr = me.MultipartReader()
	}

	l := list.New()
	l.PushBack(mr)

	return &Reader{Header{e.Header}, e, l, true}
}

// CreateReaderF reads a mail header from r and returns a new mail reader.
//
// If the message uses an unknown transfer encoding or charset, CreateReader
// returns an error that verifies message.IsUnknownCharset, but also returns a
// Reader that can be used.
//
// This defers from CreateReader by creating a error tolerant Reader
func CreateReaderF(r io.Reader) (*Reader, error) {
	e, err := message.ReadF(r)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, err
	}

	return NewReaderF(e), err
}
