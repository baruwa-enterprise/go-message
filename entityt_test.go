package message

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func testMakeMultipartF(tolerant bool) *Entity {
	var e, e1, e2 *Entity
	var h1 Header
	h1.Set("Content-Type", "text/plain")
	r1 := strings.NewReader("Text part")
	if tolerant {
		e1, _ = NewF(h1, r1)
	} else {
		e1, _ = New(h1, r1)
	}

	var h2 Header
	h2.Set("Content-Type", "text/html")
	r2 := strings.NewReader("<p>HTML part</p>")
	if tolerant {
		e2, _ = NewF(h2, r2)
	} else {
		e2, _ = New(h2, r2)
	}

	var h Header
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")
	if tolerant {
		e, _ = NewMultipartF(h, []*Entity{e1, e2})
	} else {
		e, _ = NewMultipart(h, []*Entity{e1, e2})
	}
	return e
}

func testNewMultiPartRead(t *testing.T, tolerant bool) {
	e := testMakeMultipartF(tolerant)

	if b, err := ioutil.ReadAll(e.Body); err != nil {
		t.Errorf("Expected[tolerant=%t] no error while reading multipart body, got %s", tolerant, err)
	} else if s := string(b); s != testMultipartBody {
		t.Errorf("Expected[tolerant=%t] %q as multipart body but got %q", tolerant, testMultipartBody, s)
	}
}

func TestNewMultipartF_read(t *testing.T) {
	testNewMultiPartRead(t, false)
	testNewMultiPartRead(t, true)
}

func testwalkMultipart(t *testing.T, tolerant bool) {
	e := testMakeMultipartF(tolerant)

	want := []testWalkPart{
		{
			path:      nil,
			mediaType: "multipart/alternative",
		},
		{
			path:      []int{0},
			mediaType: "text/plain",
			body:      "Text part",
		},
		{
			path:      []int{1},
			mediaType: "text/html",
			body:      "<p>HTML part</p>",
		},
	}

	got, err := walkCollect(e)
	if err != nil {
		t.Fatalf("Entity.Walk()[tolerant=%t] = %v", tolerant, err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Entity.Walk()[tolerant=%t] =\n%#v\nbut want:\n%#v", tolerant, got, want)
	}
}

func TestWalkF_multipart(t *testing.T) {
	testwalkMultipart(t, false)
	testwalkMultipart(t, true)
}
