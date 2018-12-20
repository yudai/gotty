package server

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// requestInfoReader reader of request info with JSON format
type requestInfoReader struct {
	Request *http.Request
	buf     *bytes.Buffer
}

func (this *requestInfoReader) Read(p []byte) (n int, err error) {
	if this.buf == nil {
		r := this.Request
		// request info to pass in STDIN
		info := map[string]interface{}{
			"Headers":    map[string][]string(r.Header),
			"RequestURI": r.RequestURI,
			"Referer":    r.Referer(),
			"Url": map[string]interface{}{
				"Scheme":   r.URL.Scheme,
				"Host":     r.URL.Host,
				"Path":     r.URL.Path,
				"Fragment": r.URL.Fragment,
				"Query":    r.URL.Query(),
			},
		}

		// Write data to stdin
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(info)
		this.buf = &buf
	}
	return this.buf.Read(p)
}
