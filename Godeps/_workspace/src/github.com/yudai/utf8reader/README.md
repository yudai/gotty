# UTF8Reader for Go

UTF8Reader is a simple wrapper Reader that fills the given buffer with a "tail-safe" UTF8 byte sequence.

## Tail-Safe?

Let's say you have a buffer of 7 bytes and your Reader is going to fill your buffer with a UTF8 byte sequence.

```go
buf := make([]byte, 7)
reader := strings.NewReader("いろは")

reader.Read(buf)
```

The byte length of UTF8 characters is not fixed and some characters like the examples above have 3 byte length. There are others which have a single byte, 2 byte and 4 byte length as well. This means your buffer will be sometimes filled with incomplete bytes as an Unicode character at the tail.

By `reader.Read(buf)`, your `buf` will be like below:

```go
[]byte{
	// い
	byte(0xe3), // 1
	byte(0x81), // 2
	byte(0x84), // 3
	// ろ
	byte(0xe3), // 4
	byte(0x82), // 5
	byte(0x8d), // 6
	// は (incomplete)
	byte(0xe3), // 7
}
```

The last character `は` is incomplete and the buffer is now invalid as a UTF8 string.

UTF8Reader detects incomplete bytes like above and aborts filling up the buffer in such cases.

```go
buf := make([]byte, 7)
reader := strings.NewReader("いろは")
utfReader := utf8reader.New(reader)

utfReader.Read(buf)
```

Then you will get:

```go
[]byte{
	// い
	byte(0xe3), // 1
	byte(0x81), // 2
	byte(0x84), // 3
	// ろ
	byte(0xe3), // 4
	byte(0x82), // 5
	byte(0x8d), // 6
}
```
Of course, bytes left behind will be used to fill up the buffer on next `Read()`.

## Note

UTF8Reader just checks incomplete bytes at the tail of the buffer. Even if the original byte sequence given to UTF8Reader is broken, UTF8Reader reports no errors and just fills up the buffer.
