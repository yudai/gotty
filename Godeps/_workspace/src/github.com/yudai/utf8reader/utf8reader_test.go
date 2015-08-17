package utf8reader

import (
	"testing"

	"bytes"
	"strings"
)

func TestRead(t *testing.T) {
	str := "日本語"
	or := strings.NewReader(str)
	r := New(or)

	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if bytes.Compare(buf[:n], []byte(str)) != 0 {
		t.Errorf("Failed to read bytes")
	}

	n, err = r.Read(buf)
	if err.Error() != "EOF" {
		t.Errorf("Unexpected error")
	}

	// 3 byte runes
	str = "いろはにほ"
	or = strings.NewReader(str)
	r = New(or)
	buf = make([]byte, 7) // 7 % 3 = 1

	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if n != 6 {
		t.Errorf("Read length error")
	}
	if bytes.Compare(buf[:n], []byte(str)[:6]) != 0 {
		t.Errorf("Failed to read bytes")
	}

	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if n != 6 {
		t.Errorf("Read length error")
	}
	if bytes.Compare(buf[:n], []byte(str)[6:12]) != 0 {
		t.Errorf("Failed to read bytes")
	}

	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	if n != 3 {
		t.Errorf("Read length error")
	}
	if bytes.Compare(buf[:n], []byte(str)[12:15]) != 0 {
		t.Errorf("Failed to read bytes")
	}
}

func TestReadWithSmallBuffer(t *testing.T) {
	str := "日本語"
	or := strings.NewReader(str)
	r := New(or)

	buf := make([]byte, 2) // too small
	_, err := r.Read(buf)
	if err != SmallBufferError {
		t.Errorf("Expected error were not returned")
	}
}
