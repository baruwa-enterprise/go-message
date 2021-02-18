package textproto

import (
	"bufio"
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

const (
	testHeaderMidNonFold = "Content-Disposition: inline; filename=\"6D19D_xxxxx_7D4049.zip\"\r\n" +
		"Content-Type: application/zip; x-unix-mode=0600;\r\n" +
		"name=\"6D19D_xxxxx_7D4049.zip\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n"
	testHeaderLastNonFold = "Content-Disposition: inline; filename=\"6D19D_xxxxx_7D4049.zip\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"Content-Type: application/zip; x-unix-mode=0600;\r\n" +
		"name=\"6D19D_xxxxx_7D4049.zip\"\r\n\r\n" +
		"UEsDBBQAAAAAAFebfUgAAAAAAAAAAAAAAAAFAAAAc2Nhbi9QSwMEFAAAAAAAV5t9SAAAAAAA\r\n"
	testHeaderNoSepExpect = "<html>\r\n<head>\r\n"
	testHeaderNoSep       = "Content-Transfer-Encoding: 8bit\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n" +
		testHeaderNoSepExpect
)

func testInvalidHdr(t *testing.T, tolerant bool) {
	var err error
	r := bufio.NewReader(strings.NewReader(testInvalidHeader))
	if tolerant {
		_, _, err = ReadHeaderF(r)
	} else {
		_, err = ReadHeader(r)
	}
	if err == nil {
		t.Errorf("[tolerant=%t]No error thrown", tolerant)
	}
}

func TestInvalidHeaderF(t *testing.T) {
	testInvalidHdr(t, false)
	testInvalidHdr(t, true)
}

func testReadHeader(t *testing.T, tolerant bool) {
	var err error
	var h Header
	r := bufio.NewReader(strings.NewReader(testHeader))
	if tolerant {
		h, _, err = ReadHeaderF(r)
	} else {
		h, err = ReadHeader(r)
	}
	if err != nil {
		if tolerant {
			t.Fatalf("ReadHeaderF() returned error: %v", err)
		} else {
			t.Fatalf("ReadHeader() returned error: %v", err)
		}
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Received: from example.com by example.org",
		"Received: from localhost by example.com",
		"To: Taki Tachibana <taki.tachibana@example.org>",
		"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields()[tolerant=%t] reported incorrect values: got \n%#v\n but want \n%#v", tolerant, l, want)
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != io.EOF {
		t.Errorf("Read()[tolerant=%t] didn't return EOF: %v", tolerant, err)
	}
}

func TestReadHeaderF(t *testing.T) {
	testReadHeader(t, false)
	testReadHeader(t, true)
}

func testReadHeaderWithoutBody(t *testing.T, tolerant bool) {
	var err error
	var h Header
	r := bufio.NewReader(strings.NewReader(testHeaderWithoutBody))
	if tolerant {
		h, _, err = ReadHeaderF(r)
	} else {
		h, err = ReadHeader(r)
	}
	if err != nil {
		if tolerant {
			t.Fatalf("ReadHeaderF() returned error: %v", err)
		} else {
			t.Fatalf("ReadHeader() returned error: %v", err)
		}
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Received: from example.com by example.org",
		"Received: from localhost by example.com",
		"To: Taki Tachibana <taki.tachibana@example.org>",
		"From: Mitsuha Miyamizu <mitsuha.miyamizu@example.com>",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != io.EOF {
		t.Errorf("Read() didn't return EOF: %v", err)
	}
}

func TestReadHeaderWithoutBodyF(t *testing.T) {
	testReadHeaderWithoutBody(t, false)
	testReadHeaderWithoutBody(t, true)
}

func testReadHeaderfl(t *testing.T, tolerant bool) {
	var err error
	var h Header
	r := bufio.NewReader(strings.NewReader(testLFHeader))
	if tolerant {
		h, _, err = ReadHeaderF(r)
	} else {
		h, err = ReadHeader(r)
	}
	if err != nil {
		if tolerant {
			t.Fatalf("ReadHeaderF() returned error: %v", err)
		} else {
			t.Fatalf("ReadHeader() returned error: %v", err)
		}
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"From: contact@example.org",
		"To: contact@example.org",
		"Subject: A little message, just for you",
		"Date: Wed, 11 May 2016 14:31:59 +0000",
		"Message-Id: <0000000@localhost/>",
		"Content-Type: text/plain",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	b := make([]byte, 1)
	if _, err := r.Read(b); err != io.EOF {
		t.Errorf("Read() didn't return EOF: %v", err)
	}
}

func TestReadHeaderF_lf(t *testing.T) {
	testReadHeaderfl(t, false)
	testReadHeaderfl(t, true)
}

func TestReadHeader_mid_non_fold(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(testHeaderMidNonFold))
	h, remaining, err := ReadHeaderF(r)
	if err != nil {
		t.Fatalf("readHeaderF() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Content-Disposition: inline; filename=\"6D19D_xxxxx_7D4049.zip\"",
		"Content-Type: application/zip; x-unix-mode=0600;name=\"6D19D_xxxxx_7D4049.zip\"",
		"Content-Transfer-Encoding: base64",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	if !bytes.Equal(remaining, []byte("")) {
		t.Errorf("Expected %s in the remaining slice found %s", "", remaining)
	}
}

func TestReadHeader_last_non_fold(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(testHeaderLastNonFold))
	h, remaining, err := ReadHeaderF(r)
	if err != nil {
		t.Fatalf("readHeaderF() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Content-Disposition: inline; filename=\"6D19D_xxxxx_7D4049.zip\"",
		"Content-Transfer-Encoding: base64",
		"Content-Type: application/zip; x-unix-mode=0600;name=\"6D19D_xxxxx_7D4049.zip\"",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	if !bytes.Equal(remaining, []byte("")) {
		t.Errorf("Expected %s in the remaining slice found %s", "", remaining)
	}
}

func TestReadHeader_no_sep(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(testHeaderNoSep))
	h, remaining, err := ReadHeaderF(r)
	if err != nil {
		t.Fatalf("readHeaderF() returned error: %v", err)
	}

	l := collectHeaderFields(h.Fields())
	want := []string{
		"Content-Transfer-Encoding: 8bit",
		"Content-Type: text/html; charset=\"utf-8\"",
	}
	if !reflect.DeepEqual(l, want) {
		t.Errorf("Fields() reported incorrect values: got \n%#v\n but want \n%#v", l, want)
	}

	if !bytes.Equal(remaining, []byte(testHeaderNoSepExpect)) {
		t.Errorf("Expected %s in the remaining slice found %s", testHeaderNoSepExpect, remaining)
	}
}
