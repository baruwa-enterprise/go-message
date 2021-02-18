package mail_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/emersion/go-message/mail"
)

func TestWriterF(t *testing.T) {
	var b bytes.Buffer

	var h mail.Header
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		t.Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := tw.CreatePart(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()
	tw.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "text/plain")
	ah.SetFilename("note.txt")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "I'm Mitsuha.")
	w.Close()

	mw.Close()

	testReaderF(t, &b, false)
}

func TestWriterF_singleInline(t *testing.T) {
	var b bytes.Buffer

	var h mail.Header
	h.SetSubject("Your Name")
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text part
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w, err := mw.CreateSingleInline(th)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "Who are you?")
	w.Close()

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", "text/plain")
	ah.SetFilename("note.txt")
	w, err = mw.CreateAttachment(ah)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "I'm Mitsuha.")
	w.Close()

	mw.Close()

	t.Logf("Formatted message: \n%v", b.String())

	testReaderF(t, &b, false)
}
