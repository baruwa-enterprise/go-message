package textproto

import (
	"bufio"
	"bytes"
	"io"
)

// NewMultipartReaderF creates a new multipart reader reading from r using the
// given MIME boundary.
//
// The boundary is usually obtained from the "boundary" parameter of
// the message's "Content-Type" header. Use mime.ParseMediaType to
// parse such headers.
//
// This defers from NewMultipartReader by the fact that it tolerates some
// common parsing errors
func NewMultipartReaderF(r io.Reader, boundary string) *MultipartReader {
	mr := NewMultipartReader(r, boundary)
	mr.tolerant = true
	return mr
}

func (bp *Part) populateHeadersF() error {
	header, remaining, err := ReadHeaderF(bp.mr.bufReader)
	if err == nil {
		bp.Header = header
		if len(remaining) > 0 {
			r := io.MultiReader(bytes.NewReader(remaining), bp.mr.bufReader)
			bp.mr.bufReader = bufio.NewReaderSize(&stickyErrorReader{r: r}, peekBufferSize)
		}
	}

	return err
}
