package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/compressionutils"
	bitstructs "github.com/ElwinCabrera/go-data-structs/bit-structs"
	"github.com/ElwinCabrera/go-data-structs/utils"
)

type freqInterval struct {
	start, end, width uint
	isEnd             bool
}

var ENDSYMBOL uint16 = 256

func Compress(srcData *[]byte) ([]byte, bool) {
	freqMap := compressionutils.GetSymbolFrequencyMap(srcData)
	encodedData, canCompress := EncodeWithProbabilityModel(srcData, freqMap, true)
	serializedFreqTable := serializeFrequencyTable(freqMap)

	//fmt.Printf("Serialized frequency table method 1 length: %v\n", len(serializedFreqTable))
	fmt.Printf("Serialized frequency table method 2 length: %v\n", len(serializeFrequencyTable2(freqMap)))

	compressedData := append(serializedFreqTable, encodedData...)
	return compressedData, canCompress
}

func Decompress(compressedData *[]byte) ([]byte, bool) {
	freqTable, serializedTableSize := deserializeFrequencyTable(compressedData)
	originalDataLen := uint(0)
	for _, freq := range freqTable {
		originalDataLen += uint(freq)
	}
	data := (*compressedData)[serializedTableSize:]
	return DecodeWithProbabilityModel(&data, &freqTable, originalDataLen)
}

func EncodeWithProbabilityModel(srcData *[]byte, frequencyMap *map[uint16]uint64, appendEndSymbol bool) ([]byte, bool) {
	//freqMap, _ := utils.GetSymbolProbabilityMap(srcData, false)
	symToCumFreqMap := getCumulativeFrequenciesFromFreqMap(frequencyMap, appendEndSymbol)
	return encode(srcData, symToCumFreqMap)
}

func DecodeWithProbabilityModel(encodedData *[]byte, frequencyMap *map[uint16]uint64, originalDataLen uint) ([]byte, bool) {
	symToCumFreqMap := getCumulativeFrequenciesFromFreqMap(frequencyMap, true)
	return decode(encodedData, symToCumFreqMap, originalDataLen)
}

func encode(srcData *[]byte, symToCumulativeFreq map[uint16]freqInterval) ([]byte, bool) {

	fracRepOfLow := uint16(0x0000)
	fracRepOfHigh := uint16(0xFFFF)

	total := uint(len(*srcData)) + 1 // plus end symbol
	bitSeq := bitstructs.NewDynamicBitSequence()
	underflowCounter := 0
	for i := 0; i < len(*srcData)+1; i++ {

		symStart := uint(0)
		symEnd := uint(0)
		if i < len(*srcData) {
			symStart = symToCumulativeFreq[uint16((*srcData)[i])].start
			symEnd = symToCumulativeFreq[uint16((*srcData)[i])].end
		} else {
			symStart = symToCumulativeFreq[ENDSYMBOL].start
			symEnd = symToCumulativeFreq[ENDSYMBOL].end
		}

		width := (uint(fracRepOfHigh) + 1) - uint(fracRepOfLow)
		oldFracRepLow := fracRepOfLow

		fracRepOfLow = uint16(uint(oldFracRepLow) + (width*symStart)/total) // same as fracRepOfLow  + (width * (start/total))
		fracRepOfHigh = uint16(uint(oldFracRepLow) + ((width * symEnd) / total) - 1)

		//Eliminate common bits and handle overflow. this loops until lows MSB is 0 and highs MSB is 1
		for {
			if (fracRepOfLow >> 15) == (fracRepOfHigh >> 15) {
				//MSB of low and high are both 1 (0.1....) or both 0 (0.0....)
				bit := fracRepOfLow >> 15
				bitSeq.AppendBitEnd(byte(bit))
				if bit == 0 {
					outputBitToSequenceXTimes(1, underflowCounter, &bitSeq)
				} else {
					outputBitToSequenceXTimes(0, underflowCounter, &bitSeq)
				}
				updateFracRepOfLowAndHigh(&fracRepOfLow, &fracRepOfHigh)
				underflowCounter = 0
			} else if ((fracRepOfLow>>14)&0x1) == 1 && ((fracRepOfHigh>>14)&0x1) == 0 { //check 2nd MSB (low >= 0.01... and high < 0.11...)
				//update underflow counter and fold underflow bits
				underflowCounter++
				updateFracRepOfLowAndHigh(&fracRepOfLow, &fracRepOfHigh)
				fracRepOfLow &= 0x7FFF  //make sure MSB is set to 0 for low
				fracRepOfHigh |= 0x8000 //make sure MSB is set to 1 for high

			} else { //MSB of low starts with 0 and MSB of high starts with 1
				break
			}
		}

	}

	bitSeq.AppendBitEnd(0)
	bitSeq.AppendBitEnd(1)

	unusedBitsInByte := 0
	if bitSeq.GetNumBits()%bitstructs.BYTE_LENGTH != 0 {
		unusedBitsInByte = bitstructs.BYTE_LENGTH - (bitSeq.GetNumBits() % bitstructs.BYTE_LENGTH)
		bitSeq.ExpandNumOfBitsToUseRemainingBitsInLastByte()
		for unusedBitsInByte > 0 {
			bitSeq.SetBit(bitSeq.GetNumBits()-unusedBitsInByte, true)
			unusedBitsInByte--
		}
	}

	return bitSeq.GetBitSeq(), true
}

func decode(encodedByteArray *[]byte, symToCumulativeFreq map[uint16]freqInterval, originalLen uint) ([]byte, bool) {

	fracRepOfLow := uint16(0x0000)
	fracRepOfHigh := uint16(0xFFFF)

	bitSeq := bitstructs.NewBitSequenceFromByteArray(encodedByteArray, len(*encodedByteArray)*bitstructs.BYTE_LENGTH)
	bitSeq.SetNextBitStart(0)
	fracRepOfEncodedValue := uint16(0)
	for i := 0; i < 16; i++ {
		updateFracRepOfEncodedVal(&fracRepOfEncodedValue, &bitSeq)
	}

	//fracRepOfEncodedValue := (uint16(bitSeq.GetByte(0)) << 8) | uint16(bitSeq.GetByte(1))
	//bitSeq.SetNextBitStart(16)

	totalLen := originalLen + 1 // + 1 bc of end symbol

	var decodedBuffer bytes.Buffer

	for {
		width := (uint(fracRepOfHigh) + 1) - uint(fracRepOfLow)

		// (T * (encoded - low + 1 ) -1) /((high + 1) - low)
		//(high + 1) - low = Width
		fracRepOfEncodedValueUint := uint(fracRepOfEncodedValue)
		fracRepOfLowUint := uint(fracRepOfLow)

		scaledUPValue := ((totalLen * (fracRepOfEncodedValueUint - fracRepOfLowUint + 1)) - 1) / width
		sym, interval := getSymbolWithinFreqRange(scaledUPValue, &symToCumulativeFreq)
		if interval.isEnd || sym == ENDSYMBOL {
			break
		}
		decodedBuffer.WriteByte(byte(sym))

		oldFracRepLow := fracRepOfLow
		fracRepOfLow = uint16(uint(oldFracRepLow) + (width*symToCumulativeFreq[sym].start)/totalLen) // same as fracRepOfLow  + (width * (start/total))
		fracRepOfHigh = uint16(uint(oldFracRepLow) + ((width * symToCumulativeFreq[sym].end) / totalLen) - 1)

		//Eliminate common bits and handle overflow. this loops until lows MSB is 0 and highs MSB is 1
		for {

			if (fracRepOfLow >> 15) == (fracRepOfHigh >> 15) {
				//MSB of low and high are both 1 (0.1....) or both 0 (0.0....)
				updateFracRepOfLowAndHigh(&fracRepOfLow, &fracRepOfHigh)
				updateFracRepOfEncodedVal(&fracRepOfEncodedValue, &bitSeq)
			} else if ((fracRepOfLow>>14)&0x1) == 1 && ((fracRepOfHigh>>14)&0x1) == 0 { //check 2nd MSB (low >= 0.01... and high < 0.11...)
				//remove 2nd MSB
				savedFirstBit := fracRepOfEncodedValue & 0x8000
				restOfBits := fracRepOfEncodedValue & 0x3FFF
				fracRepOfEncodedValue = savedFirstBit | (restOfBits << 1) | uint16(bitstructs.BoolToInt(bitSeq.GetNextBit()))

				updateFracRepOfLowAndHigh(&fracRepOfLow, &fracRepOfHigh)
				fracRepOfLow &= 0x7FFF  //make sure MSB is set to 0 for low
				fracRepOfHigh |= 0x8000 //make sure MSB is set to 1 for high

			} else { //MSB of low starts with 0 and MSB of high starts with 1
				break
			}

			if bitSeq.GetNextBitIdx() >= bitSeq.GetNumBits() {
				bitSeq.SetNextBitStart(bitSeq.GetNumBits() - 1)
			}

		}

	}

	return decodedBuffer.Bytes(), true
}

func serializeFrequencyTable(freqTable *map[uint16]uint64) []byte {
	// can handle a max of 255tb
	var buffer bytes.Buffer
	buffer.WriteByte(byte(len(*freqTable) - 1)) // if len is 256 -1 so that it can fit in a byte (this means that 0 is significant)
	for symbol, freq := range *freqTable {

		serializedFreqHexStr := utils.NumToHexString(freq)
		buffer.WriteByte(byte(symbol))
		for _, ch := range serializedFreqHexStr {
			buffer.WriteByte(byte(ch))
		}

		//remove zeros from most significant byte (hex str is little endian so start from 0)
		//this might not be needed because the generated hex string does not include non-significant zeros
		for i := 0; i < len(serializedFreqHexStr); i++ {
			bt := serializedFreqHexStr[i]
			if bt == '0' {
				buffer.Truncate(buffer.Len() - 1)
			} else {
				break
			}
		}
		buffer.WriteByte(0x00)

	}
	return buffer.Bytes()
}

func deserializeFrequencyTable(data *[]byte) (map[uint16]uint64, int) {
	// can handle a max of 3tb
	freqTable := make(map[uint16]uint64)

	numSymbols := int((*data)[0]) + 1
	numSymbolsSeen := 0
	serializedTableTotalSize := numSymbols*2 + 1 //  * 2
	idx := 1
	tot := uint64(0)
	// bytes, kilobytes, megabytes, gigabytes, terabytes
	unitMultipliers := []uint64{1, 1024, 1024 * 1024, 1024 * 1024 * 1024, 1024 * 1024 * 1024 * 1024}
	for numSymbolsSeen < numSymbols {
		symbol := (*data)[idx]
		idx++
		var serializedFreqHexBytes []byte

		for (*data)[idx] != 0x00 {
			serializedFreqHexBytes = append(serializedFreqHexBytes, (*data)[idx])
			serializedTableTotalSize++
			idx++
		}

		serializedFreq := utils.HexStringToInt(string(serializedFreqHexBytes))
		freq := uint64(0)
		//unitsAvailable := serializedFreq & 0x1F
		shiftAmt := 0
		for _, multiplier := range unitMultipliers {
			//unitSet := (unitsAvailable & (1 << i)) >> i
			freq += ((serializedFreq >> shiftAmt) & 0x3FF) * multiplier //* uint64(unitSet)
			shiftAmt += 10
		}

		sym, ok := freqTable[uint16(symbol)]
		if ok {
			fmt.Printf("woops seen symbol '%v' before. curr idx: %v, num syms seen so far: %v\n", sym, idx, numSymbolsSeen)
		}
		freqTable[uint16(symbol)] = freq
		tot += freq
		idx++
		numSymbolsSeen++
	}

	return freqTable, serializedTableTotalSize
}

// could useful to reduce the size of the serialized table
// if you know you will be compressing large amounts of data there might be a possibility of having a byte, kilobyte, megabyte, gigabyte, or terabyte
// as a whole round number with zero remainder
func serializeFrequencyTable2(freqTable *map[uint16]uint64) []byte {
	// can handle a max of 255tb
	var buffer bytes.Buffer
	buffer.WriteByte(byte(len(*freqTable) - 1)) // if len is 256 -1 so that it can fit in a byte (this means that 0 is significant)
	for symbol, freq := range *freqTable {

		serializedFreq := uint64(0)

		bytez := freq & 0x3FF
		kb := (freq >> 10) & 0x3FF
		mb := (freq >> 20) & 0x3FF
		gb := (freq >> 30) & 0x3FF
		tb := (freq >> 40) & 0x3FF

		units := []uint64{bytez, kb, mb, gb, tb}
		for i, unit := range units {
			if unit != 0 {
				serializedFreq |= 1 << i
			}
		}
		shiftAmt := len(units)
		for _, unit := range units {
			if unit != 0 {
				serializedFreq |= unit << shiftAmt
				shiftAmt += 10
			}
		}

		serializedFreqHexStr := utils.NumToHexString(serializedFreq)
		buffer.WriteByte(byte(symbol))
		for _, ch := range serializedFreqHexStr {
			buffer.WriteByte(byte(ch))
		}

		//remove zeros from most significant byte (hex str is little endian so start from 0)
		for i := 0; i < len(serializedFreqHexStr); i++ {
			bt := serializedFreqHexStr[i]
			if bt == '0' {
				buffer.Truncate(buffer.Len() - 1)
			} else {
				break
			}
		}
		buffer.WriteByte(0x00)

	}
	return buffer.Bytes()
}

// could useful to reduce the size of the serialized table
// If you know you will be compressing large amounts of data there might be a possibility of having a byte, kilobyte, megabyte, gigabyte, or terabyte
// as a whole round number with zero remainder
func deserializeFrequencyTable2(data *[]byte) (map[byte]uint64, int) {
	// can handle a max of 3tb
	freqTable := make(map[byte]uint64)

	numSymbols := int((*data)[0]) + 1
	numSymbolsSeen := 0
	serializedTableTotalSize := numSymbols*2 + 1 //  * 2
	idx := 1
	tot := uint64(0)
	// bytes, kilobytes, megabytes, gigabytes, terabytes
	unitMultipliers := []uint64{1, 1024, 1024 * 1024, 1024 * 1024 * 1024, 1024 * 1024 * 1024 * 1024}
	for numSymbolsSeen < numSymbols {
		symbol := (*data)[idx]
		idx++
		var serializedFreqHexBytes []byte

		for (*data)[idx] != 0x00 {
			serializedFreqHexBytes = append(serializedFreqHexBytes, (*data)[idx])
			serializedTableTotalSize++
			idx++
		}

		serializedFreq := utils.HexStringToInt(string(serializedFreqHexBytes))
		freq := uint64(0)
		unitsAvailable := serializedFreq & 0x1F
		shiftAmt := len(unitMultipliers)
		for i, multiplier := range unitMultipliers {
			unitSet := (unitsAvailable & (1 << i)) >> i
			freq += ((serializedFreq >> shiftAmt) & 0x3FF) * multiplier * uint64(unitSet)
			shiftAmt += 10
		}

		sym, ok := freqTable[symbol]
		if ok {
			fmt.Printf("woops seen symbol '%v' before. curr idx: %v, num syms seen so far: %v\n", sym, idx, numSymbolsSeen)
		}
		freqTable[symbol] = freq
		tot += freq
		idx++
		numSymbolsSeen++
	}

	return freqTable, serializedTableTotalSize
}

// Helpers
func outputBitToSequenceXTimes(bit0or1, numTimes int, bitSequence *bitstructs.BitSequence) {
	for i := 0; i < numTimes; i++ {
		bitSequence.AppendBitEnd(byte(bit0or1))
	}
}

func updateFracRepOfLowAndHigh(fracRepLow, fracRepHigh *uint16) {
	*fracRepLow <<= 1
	*fracRepHigh = (*fracRepHigh << 1) | 0x1
}

func updateFracRepOfEncodedVal(fracRepOfEncoded *uint16, bitSequence *bitstructs.BitSequence) {
	*fracRepOfEncoded <<= 1
	*fracRepOfEncoded |= uint16(bitstructs.BoolToInt(bitSequence.GetNextBit()))
}

func getCumulativeFrequenciesFromFreqMap(frequencyMap *map[uint16]uint64, appendEndSymbol bool) map[uint16]freqInterval {
	symToCumulativeFreq := make(map[uint16]freqInterval)
	prevEnd := uint(0)
	sortedMapKeys := compressionutils.GetArrayOfSortedMapKeys(frequencyMap) // we need to guarantee order for because each interval needs to be divided up the same way for a static model
	for _, key := range sortedMapKeys {
		sym := uint16(key)
		symFreq := uint((*frequencyMap)[sym])
		symToCumulativeFreq[sym] = freqInterval{prevEnd, prevEnd + symFreq, symFreq, false}
		prevEnd += symFreq
	}
	if appendEndSymbol {
		symToCumulativeFreq[ENDSYMBOL] = freqInterval{prevEnd, prevEnd + 1, 1, true}
	}
	return symToCumulativeFreq
}

func getSymbolWithinFreqRange(value uint, symToCumulativeFreq *map[uint16]freqInterval) (uint16, freqInterval) {
	for sym, inter := range *symToCumulativeFreq {
		if value >= inter.start && value < inter.end {
			return sym, inter
		}
	}
	//fmt.Printf("Error!!\n")
	return 0, freqInterval{}
}
