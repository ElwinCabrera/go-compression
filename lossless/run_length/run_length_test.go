package run_length

import (
	"fmt"
	"testing"
)

func TestRunLengthEncode(t *testing.T) {
	data := []byte("WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWWWWWWWWWWWWBWWWWWWWWWWWWWW")
	//expected 12W 1B 12W 3B 24W 1B 14W
	compressedData := RunLengthEncode(data)

	fmt.Println(string(compressedData))
	uncompressedData := RunLengthDecode(compressedData)

	if !verifyArraysEqual(data, uncompressedData) {
		t.Errorf("compressed data does not match uncompressed data")
	}
}

func verifyArraysEqual(a1 []byte, a2 []byte) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i := range a1 {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func verifyEncodedData(t *testing.T, data []byte) {
	freqMap := make(map[byte]int)
	for _, b := range data {
		freqMap[b]++
	}

}
