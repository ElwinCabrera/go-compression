package run_length

import "bytes"

func RunLengthEncode(data []byte) []byte {
	var buffer bytes.Buffer
	runLength := uint8(1)
	for i := 0; i < len(data); i++ {

		if i+1 < len(data) && data[i] == data[i+1] {
			runLength++
			if runLength == 0xFF {
				buffer.WriteByte(0xFF)
				buffer.WriteByte(runLength)
				runLength = 1
			}
		} else {
			buffer.WriteByte(runLength)
			if i == 0 {
				buffer.WriteByte(data[0])
			} else {
				buffer.WriteByte(data[i])
			}
			runLength = 1
		}
	}
	return buffer.Bytes()
}

func RunLengthDecode(encodedData []byte) []byte {
	if len(encodedData) < 2 {
		return nil
	}
	var buffer bytes.Buffer
	for i := 0; i < len(encodedData); i += 2 {
		for j := uint8(0); j < encodedData[i]; j++ {
			buffer.WriteByte(encodedData[i+1])
		}
	}
	return buffer.Bytes()
}
