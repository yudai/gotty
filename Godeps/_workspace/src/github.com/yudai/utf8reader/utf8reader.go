package utf8reader

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

var SmallBufferError = errors.New("Buffer size must be larger than utf8.UTFMax.")

type UTF8Reader struct {
	reader io.Reader
	buffer *bytes.Buffer
}

func New(reader io.Reader) *UTF8Reader {
	return &UTF8Reader{
		reader: reader,
		buffer: bytes.NewBuffer(make([]byte, 0)),
	}
}

func (r *UTF8Reader) Read(p []byte) (n int, err error) {
	size := 0

	if cap(p) < utf8.UTFMax {
		return size, SmallBufferError
	}

	if r.buffer.Len() > 0 {
		n, err = r.buffer.Read(p)
		size += n
		if err != nil {
			return size, err
		}
	}

	n, err = r.reader.Read(p[size:])
	size += n
	if err != nil {
		return size, err
	}

	leftOver := 0
	for ; leftOver < utf8.UTFMax && size-leftOver > 0; leftOver++ {
		rune, _ := utf8.DecodeLastRune(p[:size-leftOver])
		if rune != utf8.RuneError {
			break
		}
	}

	r.buffer.Write(p[size-leftOver : size])

	return size - leftOver, nil
}
