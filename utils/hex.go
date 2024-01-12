package utils

import (
	"encoding/hex"
	"strings"
)

func Bytes2Hex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func Hex2Bytes(str string) []byte {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}

	if len(str)%2 == 1 {
		str = "0" + str
	}

	h, _ := hex.DecodeString(str)
	return h
}

// with prefix '0x'
func Bytes2HexP(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

// TrimLeftZeroes returns a subslice of s without leading zeroes
func TrimLeftZeroes(s []byte) []byte {
	idx := 0
	for ; idx < len(s); idx++ {
		if s[idx] != 0 {
			break
		}
	}
	return s[idx:]
}

// TrimRightZeroes returns a subslice of s without trailing zeroes
func TrimRightZeroes(s []byte) []byte {
	idx := len(s)
	for ; idx > 0; idx-- {
		if s[idx-1] != 0 {
			break
		}
	}
	return s[:idx]
}
