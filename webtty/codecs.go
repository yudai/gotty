package webtty

type Decoder interface {
	Decode(dst, src []byte) (int, error)
}

type Encoder interface {
	Encode(dst, src []byte) (int, error)
}

type NullCodec struct{}

func (NullCodec) Encode(dst, src []byte) (int, error) {
	return copy(dst, src), nil
}

func (NullCodec) Decode(dst, src []byte) (int, error) {
	return copy(dst, src), nil
}
