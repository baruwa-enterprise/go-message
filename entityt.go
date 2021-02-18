package message

import (
	"bufio"
	"bytes"
	"io"
	"math"

	"github.com/emersion/go-message/textproto"
)

// NewF makes a new message with the provided header and body. The entity's
// transfer encoding and charset are automatically decoded to UTF-8.
//
// If the message uses an unknown transfer encoding or charset, New returns an
// error that verifies IsUnknownCharset, but also returns an Entity that can
// be read.
//
// This defers from New by creating a error tolerant Entity
func NewF(header Header, body io.Reader) (*Entity, error) {
	e, err := New(header, body)
	e.tolerant = true
	return e, err
}

// NewMultipartF makes a new multipart message with the provided header and
// parts. The Content-Type header must begin with "multipart/".
//
// If the message uses an unknown transfer encoding, NewMultipart returns an
// error that verifies IsUnknownCharset, but also returns an Entity that can
// be read.
//
// This defers from NewMultipart by creating a error tolerant Entity
func NewMultipartF(header Header, parts []*Entity) (*Entity, error) {
	r := &multipartBody{
		header: header,
		parts:  parts,
	}

	return NewF(header, r)
}

// ReadF reads a message from r. The message's encoding and charset are
// automatically decoded to raw UTF-8. Note that this function only reads the
// message header.
//
// If the message uses an unknown transfer encoding or charset, Read returns an
// error that verifies IsUnknownCharset or IsUnknownEncoding, but also returns
// an Entity that can be read.
//
// This defers from Read by creating a error tolerant Entity
func ReadF(r io.Reader) (*Entity, error) {
	lr := &limitedReader{R: r, N: maxHeaderBytes}
	br := bufio.NewReader(lr)

	h, remaining, err := textproto.ReadHeaderF(br)
	if err != nil {
		return nil, err
	}

	lr.N = math.MaxInt64

	if len(remaining) > 0 {
		br := io.MultiReader(&limitedReader{R: bytes.NewReader(remaining), N: maxHeaderBytes}, br)
		return NewF(Header{h}, br)
	}

	return NewF(Header{h}, br)
}
