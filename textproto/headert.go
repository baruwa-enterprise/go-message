package textproto

import (
	"bufio"
	"bytes"
	"fmt"
	"net/textproto"
	"strings"
)

// ReadHeaderF reads a MIME header from r. The header is a sequence of possibly
// continued Key: Value lines ending in a blank line.
//
// To avoid denial of service attacks, the provided bufio.Reader should be
// reading from an io.LimitedReader or a similar Reader to bound the size of
// headers.
//
// This defers from ReadHeader by handling malformed errors
func ReadHeaderF(r *bufio.Reader) (Header, []byte, error) {
	fs := make([]*headerField, 0, 32)

	var (
		remaining []byte
		th        *headerField
		kv, tkv   []byte
		err       error
	)

	// The first line cannot start with a leading space.
	if buf, err := r.Peek(1); err == nil && isSpace(buf[0]) {
		line, err := readLineSlice(r, nil)
		if err != nil {
			return newHeader(fs), remaining, err
		}

		return newHeader(fs), remaining, fmt.Errorf("message: malformed MIME header initial line: %v", string(line))
	}

	for {
		if kv, err = readContinuedLineSlice(r); len(kv) == 0 {
			return newHeader(fs), remaining, err
		}

		// Key ends at first colon; should not have trailing spaces but they
		// appear in the wild, violating specs, so we remove them if present.
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			// check the next line (readahead)
			origErr := err
			if tkv, err = readContinuedLineSlice(r); err != nil {
				// readahead failed so lets return the original failure
				return newHeader(fs), remaining, fmt.Errorf("message: malformed MIME header line: %v", string(kv))
			}

			if len(tkv) == 0 {
				// readahead found the separator with the body so append
				// the previous value to the last header value and then
				// return remaining byte slice will be empty as our
				// readahead did not encrouch on the body
				th, fs = fs[len(fs)-1], fs[:len(fs)-1]
				appendHdrVal(th, &fs, kv)
				return newHeader(fs), remaining, origErr
			}
			// we found something so lets check if it is a header
			ii := bytes.IndexByte(tkv, ':')
			if ii < 0 {
				// readahead found a non header so it shoud be
				// part of the body append the original line which caused
				// the error as well as the readahead line into the remaining
				// byte slice. This will be available to read as the body
				remaining = append(remaining, kv...)
				remaining = append(remaining, tkv...)
				return newHeader(fs), remaining, origErr
			}
			// readahead found a header set the values from the (readahead) to
			// the normal read values after appending the value to the previous
			// header
			th, fs = fs[len(fs)-1], fs[:len(fs)-1]
			appendHdrVal(th, &fs, kv)
			kv = tkv
			i = ii
		}

		keyBytes := trim(kv[:i])

		// Verify that there are no invalid characters in the header key.
		// See RFC 5322 Section 2.2
		for _, c := range keyBytes {
			if !validHeaderKeyByte(c) {
				return newHeader(fs), remaining, fmt.Errorf("message: malformed MIME header key: %v", string(keyBytes))
			}
		}

		key := textproto.CanonicalMIMEHeaderKey(string(keyBytes))

		// As per RFC 7230 field-name is a token, tokens consist of one or more
		// chars. We could return a an error here, but better to be liberal in
		// what we accept, so if we get an empty key, skip it.
		if key == "" {
			continue
		}

		i++ // skip colon
		v := kv[i:]

		value := trimAroundNewlines(v)
		fs = append(fs, newHeaderField(key, value, kv))

		if err != nil {
			return newHeader(fs), remaining, err
		}
	}
}

func appendHdrVal(th *headerField, fs *[]*headerField, kv []byte) {
	var (
		tb strings.Builder
		tv string
	)

	tv = trimAroundNewlines(kv)
	tb.WriteString(th.v)
	tb.WriteString(tv)
	th.v = tb.String()
	*fs = append(*fs, th)
}
