package util

import (
	"bytes"
	"compress/flate"
	"io"
	"strings"
)

func CompressString(in string) string {
	buf := new(bytes.Buffer)
	compressor, _ := flate.NewWriter(buf, 9)
	compressor.Write([]byte(in))
	compressor.Close()
	return buf.String()
}

func DecompressString(in string) string {
	buf := new(bytes.Buffer)
	decompressor := flate.NewReader(strings.NewReader(in))
	io.Copy(buf, decompressor)
	decompressor.Close()
	return buf.String()
}

func Compress(in []byte) []byte {
	buf := new(bytes.Buffer)
	compressor, _ := flate.NewWriter(buf, 9)
	compressor.Write(in)
	compressor.Close()
	return buf.Bytes()
}

func Decompress(in []byte) []byte {
	buf := new(bytes.Buffer)
	decompressor := flate.NewReader(bytes.NewReader(in))
	io.Copy(buf, decompressor)
	decompressor.Close()
	return buf.Bytes()
}
