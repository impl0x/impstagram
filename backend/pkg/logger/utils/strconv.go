package utils

import "bytes"

func WriteUint16(b *bytes.Buffer, u uint16) {
	if u == 0 {
		b.WriteByte('0')
		return
	}
	var buf [5]byte
	pos := 5

	for u > 0 {
		pos--
		buf[pos] = '0' + byte(u%10)
		u /= 10
	}
	b.Write(buf[pos:])
}

func WriteUint8(b *bytes.Buffer, u uint8) {
	if u == 0 {
		b.WriteByte('0')
		return
	}
	var buf [3]byte
	pos := 3
	for u > 0 {
		pos--
		buf[pos] = '0' + (u % 10)
		u /= 10
	}
	b.Write(buf[pos:])
}
