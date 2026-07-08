package font

import (
	"encoding/binary"
	"io"
)

// Manual big-endian helpers avoid encoding/binary reflection in font loops (PERF-107).

func putU16BE(dst *[]byte, v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	*dst = append(*dst, b[:]...)
}

func putU32BE(dst *[]byte, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	*dst = append(*dst, b[:]...)
}

func putI16BE(dst *[]byte, v int16) {
	putU16BE(dst, uint16(v))
}

func readU16BE(b []byte, off int) (uint16, int, error) {
	if off+2 > len(b) {
		return 0, off, io.ErrUnexpectedEOF
	}
	return binary.BigEndian.Uint16(b[off : off+2]), off + 2, nil
}

func readU32BE(b []byte, off int) (uint32, int, error) {
	if off+4 > len(b) {
		return 0, off, io.ErrUnexpectedEOF
	}
	return binary.BigEndian.Uint32(b[off : off+4]), off + 4, nil
}

func readI16BE(b []byte, off int) (int16, int, error) {
	u, off, err := readU16BE(b, off)
	return int16(u), off, err
}

// writeU16BE writes a big-endian uint16 into a growable buffer without reflection.
func writeU16BE(buf *[]byte, v uint16) {
	putU16BE(buf, v)
}

func writeU32BE(buf *[]byte, v uint32) {
	putU32BE(buf, v)
}
