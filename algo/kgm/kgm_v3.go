package kgm

import (
	"crypto/md5"
	"fmt"
	"io"
)

// kgmCryptoV3 is kgm file crypto v3
type kgmCryptoV3 struct {
	slotBox []byte
	fileBox []byte

	rd     io.Reader
	offset int
}

var kgmV3Slot2Key = map[uint32][]byte{
	1: {0x6C, 0x2C, 0x2F, 0x27},
}

func newKgmCryptoV3(header *Header, body io.Reader) (io.Reader, error) {
	c := &kgmCryptoV3{rd: body}

	slotKey, ok := kgmV3Slot2Key[header.CryptoSlot]
	if !ok {
		return nil, fmt.Errorf("kgm3: unknown crypto slot %d", header.CryptoSlot)
	}
	c.slotBox = kugouMD5(slotKey)

	c.fileBox = append(kugouMD5(header.CryptoKey), 0x6b)

	return c, nil
}

func (d *kgmCryptoV3) Read(buf []byte) (int, error) {
	n, err := d.rd.Read(buf)
	if n > 0 {
		d.decrypt(buf[:n], d.offset)
		d.offset += n
	}
	return n, err
}

func (d *kgmCryptoV3) decrypt(b []byte, offset int) {
	for i := 0; i < len(b); i++ {
		b[i] ^= d.fileBox[(offset+i)%len(d.fileBox)]
		b[i] ^= b[i] << 4
		b[i] ^= d.slotBox[(offset+i)%len(d.slotBox)]
		b[i] ^= xorCollapseUint32(uint32(offset + i))
	}
}

func xorCollapseUint32(i uint32) byte {
	return byte(i) ^ byte(i>>8) ^ byte(i>>16) ^ byte(i>>24)
}

func kugouMD5(b []byte) []byte {
	digest := md5.Sum(b)
	ret := make([]byte, 16)
	for i := 0; i < md5.Size; i += 2 {
		ret[i] = digest[14-i]
		ret[i+1] = digest[14-i+1]
	}
	return ret
}
