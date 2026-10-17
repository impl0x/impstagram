package utils

import "bytes"

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func WriteInteger[T Integer](b *bytes.Buffer, n T) {
	if n < 0 {
		b.WriteByte('-')
		writeUint64(b, uint64(-(n+1))+1)
		return
	}

	writeUint64(b, uint64(n))
}

func writeUint64(b *bytes.Buffer, n uint64) {
	var buf [20]byte // max length of max uint64
	pos := len(buf)

	if n == 0 {
		b.WriteByte('0')
		return
	}

	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}

	b.Write(buf[pos:])
}
