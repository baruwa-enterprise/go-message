package mail_test

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/emersion/go-message/mail"
)

func testreader_nonMultipart(t *testing.T, tolerant bool) {
	var mr *mail.Reader
	var err error

	s := "Subject: Your Name\r\n" +
		"\r\n" +
		"Who are you?"

	if tolerant {
		mr, err = mail.CreateReaderF(strings.NewReader(s))
	} else {
		mr, err = mail.CreateReader(strings.NewReader(s))
	}

	if err != nil {
		t.Fatalf("Expected[tolerant=%t] no error while creating reader, got: %s", tolerant, err)
	}
	defer mr.Close()

	p, err := mr.NextPart()
	if err != nil {
		t.Fatalf("Expected[tolerant=%t] no error while reading part, got: %s", tolerant, err)
	}

	if _, ok := p.Header.(*mail.InlineHeader); !ok {
		t.Fatalf("Expected[tolerant=%t] a InlineHeader, but got a %T", tolerant, p.Header)
	}

	expectedBody := "Who are you?"
	if b, err := ioutil.ReadAll(p.Body); err != nil {
		t.Errorf("Expected[tolerant=%t] no error while reading part body, but got: %s", tolerant, err)
	} else if string(b) != expectedBody {
		t.Errorf("Expected[tolerant=%t] part body to be:\n%v\nbut got:\n%v", tolerant, expectedBody, string(b))
	}

	if _, err := mr.NextPart(); err != io.EOF {
		t.Fatalf("Expected[tolerant=%t] io.EOF while reading part, but got: %s", tolerant, err)
	}
}

func TestReaderF_nonMultipart(t *testing.T) {
	testreader_nonMultipart(t, false)
	testreader_nonMultipart(t, true)
}

func testreader_closeImmediately(t *testing.T, tolerant bool) {
	var mr *mail.Reader
	var err error

	s := "Content-Type: text/plain\r\n" +
		"\r\n" +
		"Who are you?"

	if tolerant {
		mr, err = mail.CreateReaderF(strings.NewReader(s))
	} else {
		mr, err = mail.CreateReader(strings.NewReader(s))
	}

	if err != nil {
		t.Fatalf("Expected[tolerant=%t] no error while creating reader, got: %s", tolerant, err)
	}

	mr.Close()

	if _, err := mr.NextPart(); err != io.EOF {
		t.Fatalf("Expected[tolerant=%t] io.EOF while reading part, but got: %s", tolerant, err)
	}
}

func TestReaderF_closeImmediately(t *testing.T) {
	testreader_closeImmediately(t, false)
	testreader_closeImmediately(t, true)
}

func testreader_nested(t *testing.T, tolerant bool) {
	var mr *mail.Reader
	var err error

	r := strings.NewReader(nestedMailString)

	if tolerant {
		mr, err = mail.CreateReaderF(r)
	} else {
		mr, err = mail.CreateReader(r)
	}
	if err != nil {
		if tolerant {
			t.Fatalf("mail.CreateReaderF(r) = %v", err)
		} else {
			t.Fatalf("mail.CreateReader(r) = %v", err)
		}
	}
	defer mr.Close()

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		switch i {
		case 0:
			_, ok := p.Header.(*mail.InlineHeader)
			if !ok {
				t.Fatalf("Expected[tolerant=%t] a InlineHeader, but got a %T", tolerant, p.Header)
			}

			expectedBody := "I forgot."
			if b, err := ioutil.ReadAll(p.Body); err != nil {
				t.Errorf("Expected[tolerant=%t] no error while reading part body, but got: %s", tolerant, err)
			} else if string(b) != expectedBody {
				t.Errorf("Expected[tolerant=%t] part body to be:\n%v\nbut got:\n%v", tolerant, expectedBody, string(b))
			}
		case 1:
			_, ok := p.Header.(*mail.AttachmentHeader)
			if !ok {
				t.Fatalf("Expected[tolerant=%t] an AttachmentHeader, but got a %T", tolerant, p.Header)
			}

			testReaderF(t, p.Body, false)
		}

		i++
	}
}

func TestReaderF_nested(t *testing.T) {
	testreader_nested(t, false)
	testreader_nested(t, true)
}

func testReaderF(t *testing.T, r io.Reader, tolerant bool) {
	var mr *mail.Reader
	var err error

	if tolerant {
		mr, err = mail.CreateReaderF(r)
	} else {
		mr, err = mail.CreateReader(r)
	}
	if err != nil {
		if tolerant {
			t.Fatalf("mail.CreateReaderF(r) = %v", err)
		} else {
			t.Fatalf("mail.CreateReader(r) = %v", err)
		}
	}
	defer mr.Close()

	wantSubject := "Your Name"
	subject, err := mr.Header.Subject()
	if err != nil {
		t.Errorf("mr.Header.Subject() = %v", err)
	} else if subject != wantSubject {
		t.Errorf("mr.Header.Subject() = '%v', want '%v'", subject, wantSubject)
	}

	i := 0
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		var expectedBody string
		switch i {
		case 0:
			h, ok := p.Header.(*mail.InlineHeader)
			if !ok {
				t.Fatalf("Expected a InlineHeader, but got a %T", p.Header)
			}

			if mediaType, _, _ := h.ContentType(); mediaType != "text/plain" {
				t.Errorf("Expected a plaintext part, not an HTML part")
			}

			expectedBody = "Who are you?"
		case 1:
			h, ok := p.Header.(*mail.AttachmentHeader)
			if !ok {
				t.Fatalf("Expected an AttachmentHeader, but got a %T", p.Header)
			}

			if filename, err := h.Filename(); err != nil {
				t.Error("Expected no error while parsing filename, but got:", err)
			} else if filename != "note.txt" {
				t.Errorf("Expected filename to be %q but got %q", "note.txt", filename)
			}

			expectedBody = "I'm Mitsuha."
		}

		if b, err := ioutil.ReadAll(p.Body); err != nil {
			t.Error("Expected no error while reading part body, but got:", err)
		} else if string(b) != expectedBody {
			t.Errorf("Expected part body to be:\n%v\nbut got:\n%v", expectedBody, string(b))
		}

		i++
	}

	if i != 2 {
		t.Errorf("Expected exactly two parts but got %v", i)
	}
}

func TestReaderF(t *testing.T) {
	testReaderF(t, strings.NewReader(mailString), false)
	testReaderF(t, strings.NewReader(mailString), true)
}
