package huffman

import (
	"bytes"
	utils2 "github.com/ElwinCabrera/go-compression/compressionutils"
	"github.com/ElwinCabrera/go-data-structs/bit-structs"
	"github.com/ElwinCabrera/go-data-structs/trees"
	"github.com/ElwinCabrera/go-data-structs/utils"
)

func Compress(dataToCompress *[]byte) ([]byte, bool) {

	canCompress := true

	//Get the frequency each byte appears in the data we want to compress
	freqMap := utils2.GetSymbolFrequencyMap(dataToCompress)

	// Create a compression tree and generate the corresponding compression code for each unique byte with the weights being the frequency that each byte occurs
	ht := trees.NewHuffmanTreeFromFrequencyMap(*freqMap)
	huffmanCodes := ht.GetHuffmanCodes()

	//figure out the total bit length of the compressed data and create a new bit sequence to store the soon-to-be compressed data
	compressedBitLen := 0
	for _, bt := range *dataToCompress {
		compressedBitLen += huffmanCodes[uint16(bt)].GetNumBits()
	}
	bitSequenceOfCompressedData := bitstructs.NewBitSequence(compressedBitLen)

	//For each character in the original uncompressed data append its corresponding compression code to the bit sequence
	compressedDataBitIdx := 0
	for _, bt := range *dataToCompress {
		huffmanCodeBitSeq := huffmanCodes[uint16(bt)]
		bitIdx := huffmanCodeBitSeq.GetNumBits() - 1
		for bitIdx >= 0 {
			bitSequenceOfCompressedData.SetBit(compressedDataBitIdx, huffmanCodeBitSeq.GetBit(bitIdx))
			bitIdx--
			compressedDataBitIdx++
		}
	}

	//Finally generate a serialized version of the generated compression code to be sent
	//with the compressed data as these codes are the key needed for decompression.
	serializedHuffmanCodes := serializeHuffmanCodes(huffmanCodes)
	//compressedByteLen := len(serializedHuffmanCodes) + bitSequenceOfCompressedData.GetBytesAllocated()

	//Create a buffer to hold the serialized compression codes + the compressed data and start creating our end result
	var compressedData bytes.Buffer
	for _, bt := range serializedHuffmanCodes {
		compressedData.WriteByte(bt)
	}
	for i := 0; i < bitSequenceOfCompressedData.GetBytesAllocated(); i++ {
		compressedData.WriteByte(bitSequenceOfCompressedData.GetByte(i))
	}

	if compressedData.Len() >= len(*dataToCompress) {
		//If this is the case our compression resulted in a size bigger than the length of our original data
		//this happens because we need to also append the compression codes with the compressed data which can sometimes add
		//a lot more overhead to the total size of the compressed string.
		//This is especially true for smaller data that we want to compress since in that case the compression code overhead
		//will almost always be longer than the compressed data or even the original data.
		//One way we can reduce this overhead is to also compression compress the serialized compression codes agan and again until it cant be compressed anymore
		canCompress = false
		//return *dataToCompress, canCompress
	}

	numOfTrailingZerosToIgnore := uint8(0)
	if compressedBitLen%bitstructs.BYTE_LENGTH != 0 {
		numOfTrailingZerosToIgnore = uint8(bitstructs.BYTE_LENGTH - (compressedBitLen % bitstructs.BYTE_LENGTH))
	}
	compressedData.WriteByte(numOfTrailingZerosToIgnore)
	return compressedData.Bytes(), canCompress

}

func Decompress(data *[]byte) *[]byte {
	//recreate the original compression codes
	originalHuffmanCodes, serializedLen := deSerializeHuffmanCodesFromByteArray(data)
	//serializedHuffmanCodeByteArr := (*data)[:serializedLen]

	compressedData := (*data)[serializedLen : len(*data)-1]

	trailingZeroBitsToRemove := (*data)[len(*data)-1]
	bitLen := (bitstructs.BYTE_LENGTH * len(compressedData)) - int(trailingZeroBitsToRemove)

	bitSequenceOfCompressedData := bitstructs.NewBitSequenceFromByteArray(&compressedData, bitLen)

	hTree := trees.NewHuffmanTreeFromHuffmanCodes(originalHuffmanCodes)
	uncompressedData := hTree.DecodeBitSequence(&bitSequenceOfCompressedData)

	return uncompressedData
}

func serializeHuffmanCodes(hc map[uint16]bitstructs.BitSequence) []byte {
	//Will serialize as
	//<char><bit_len><code-in-hex>_<char><bit_len>...<END (2 0x00 bytes)>
	//1 byte  1 byte      X bytes					  2 0x00 bytes
	//Note: we use a bit length of 1 byte. if there's ever a case where we have a tree with depth > 255 this will fail causing incorrect data when decompressing most systems are 32 or 64 bits so a tree depth above that seems like unlikely so i think we are safe with one byte here.

	var buf bytes.Buffer

	//byteToStringMap := make(map[byte]string)

	for elem, bs := range hc {
		buf.WriteByte(byte(elem))
		buf.WriteByte(byte(bs.GetNumBits()))
		num := bs.GetXBytes(8)
		str := utils.NumToHexString(num)
		buf.Write([]byte(str))
		buf.WriteByte('_')
	}
	buf.Truncate(buf.Len() - 1)

	// signals end of compression codes use two bytes because if the uncompressed data has a zero that we need to compression encode
	// a single 0 byte to signal the end of the serialized string will cause problems as when deserializing it will stop at the
	//first null byte. Since compression codes are unique checking for a second null byte right after will ensure that we are
	//at the end of the serialized string when we are trying to deserialize, this also works because according to our
	//serialization rule there will never be a case where two null bytes appear back to back.
	//One drawback of this of course if that we are adding more (needed) overhead to our final compressed string :(
	buf.WriteByte(0x00)
	buf.WriteByte(0x00)

	return buf.Bytes()
}

func deSerializeHuffmanCodesFromByteArray(data *[]byte) (map[byte]bitstructs.BitSequence, int) {

	originalHuffmanCodes := make(map[byte]bitstructs.BitSequence)
	if len(*data) < 5 {
		return originalHuffmanCodes, 0
	}

	serializedHuffmanCodeLen := 1
	for i, bt := range *data {
		serializedHuffmanCodeLen++
		if bt == 0x00 && i+1 < len(*data) && (*data)[i+1] == 0x00 {
			break
		}
	}

	//serializedHuffmanCodes := (*data)[:serializedHuffmanCodeLen]
	idx := 0
	for idx < serializedHuffmanCodeLen && !isEndOfSerializedHuffmanCodes(data, idx, serializedHuffmanCodeLen) {
		ch := (*data)[idx]
		idx++
		bitSeq := bitstructs.NewBitSequence(int((*data)[idx]))
		idx++
		hexStr := ""
		for (*data)[idx] != '_' && !isEndOfSerializedHuffmanCodes(data, idx, serializedHuffmanCodeLen) {
			hexStr += string((*data)[idx])
			idx++
		}
		if hexStr != "" {
			num := utils.HexStringToInt(hexStr)
			bitSeq.SetBitsFromNum(0, uint64(num))
		}
		if (*data)[idx] == '_' {
			idx++
		}

		originalHuffmanCodes[ch] = bitSeq

	}
	return originalHuffmanCodes, serializedHuffmanCodeLen
}

func isEndOfSerializedHuffmanCodes(data *[]byte, idx int, endIdx int) bool {
	return (idx < endIdx && (*data)[idx] == 0x00) && (idx+1 < endIdx && (*data)[idx+1] == 0x00)
}
