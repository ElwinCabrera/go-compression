package arithmeticcoding

//This is the finite version because we are using the infinite version of the algorithm
//and since our computer only has a finite number of bits to represent probabilities aka floats
//this can only compress a finite number of characters (~32 characters if using 64 bit floating point number)

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/compressionutils"
	bitstructs "github.com/ElwinCabrera/go-data-structs/bit-structs"
	"math"
)

type probInterval struct {
	start, end, width float64
}

func finiteCompress(srcData *[]byte) ([]byte, bool) {
	freqMap := compressionutils.GetSymbolFrequencyMap(srcData)
	probabilityMap := compressionutils.GetSymbolProbMapFromFreqMap(freqMap, len(*srcData))
	return finiteEncodeWithProbabilityModel(srcData, probabilityMap)
}

func finiteDecompress(encodedData *[]byte) []byte {
	//TODO: Need to find a way to send the probability map
	probabilityMap := map[uint16]float64{}
	return finiteDecodeWithProbabilityModel(encodedData, &probabilityMap)
}

func finiteEncodeWithProbabilityModel(srcData *[]byte, probabilityMap *map[uint16]float64) ([]byte, bool) {
	//_, canCompress := canCompressData(*probabilityMap, float64(len(*srcData)))
	//if !canCompress {
	//	return []byte{}, false
	//}
	symToCumProb := getCumulativeProbabilitiesFromProbMap(probabilityMap)
	return finiteEncode(srcData, symToCumProb)
}

func finiteDecodeWithProbabilityModel(encodedBytes *[]byte, probabilityMap *map[uint16]float64) []byte {
	cm := getCumulativeProbabilitiesFromProbMap(probabilityMap)
	res, _ := finiteDecode(encodedBytes, cm)
	return res
}

func finiteEncode(srcData *[]byte, symToCumulativeProb map[uint16]probInterval) ([]byte, bool) {

	end := 1.0
	start := 0.0

	for i := 0; i < len(*srcData)+1; i++ {
		sym := uint16(0)
		if i >= len(*srcData) {
			sym = ENDSYMBOL
		} else {
			sym = uint16((*srcData)[i])
		}

		width := end - start
		end = start + (width * symToCumulativeProb[sym].end)
		start = start + (width * symToCumulativeProb[sym].start)

		if start == end {
			fmt.Printf("Reached maximum number of symbols that can be encoded using 64-bit floating point math\n")
		}
	}
	fmt.Printf("lowbound: %v, upperbound: %v, width: %v\n", start, end, end-start)
	//encodedData := utils.GetMidPointProtectingAgainstOverflow(start, end)

	bitSeq := bitstructs.NewDynamicBitSequence()
	for {
		if start > 0.5 && end > 0.5 {
			bitSeq.AppendBitEnd(1)
			start = (start / 0.5) - 1
			end = (end / 0.5) - 1
		} else if start < 0.5 && end < 0.5 {
			bitSeq.AppendBitEnd(0)
			start = start / 0.5
			end = end / 0.5
		} else if start < 0.5 && end > 0.5 {
			start = start / 0.5
			end = (end / 0.5) - 1
		} else {
			break
		}

	}
	setBit := true
	if start > 0.5 {
		bitSeq.AppendBitEnd(1)
	} else {
		bitSeq.AppendBitEnd(0)
		setBit = false
	}
	fmt.Printf("Binary representation: 0.%v\n", bitSeq)

	unusedBitsInByte := 0
	if bitSeq.GetNumBits()%bitstructs.BYTE_LENGTH != 0 {
		unusedBitsInByte = bitstructs.BYTE_LENGTH - (bitSeq.GetNumBits() % bitstructs.BYTE_LENGTH)
		bitSeq.ExpandNumOfBitsToUseRemainingBitsInLastByte()
		for unusedBitsInByte > 0 {
			bitSeq.SetBit(bitSeq.GetNumBits()-unusedBitsInByte, setBit)
			unusedBitsInByte--
		}
	}

	fmt.Printf("Binary representation (after adding padding): 0.%v\n", bitSeq)

	return bitSeq.GetBitSeq(), true
}

func finiteDecode(encodedBytes *[]byte, symToCumulativeProb map[uint16]probInterval) ([]byte, bool) {

	encodedDataAsNum := 0.0
	bitLen := len(*encodedBytes) * bitstructs.BYTE_LENGTH
	bitSeq := bitstructs.NewBitSequenceFromByteArray(encodedBytes, bitLen)
	for i := 0; i < bitSeq.GetNumBits(); i++ {
		bit := bitstructs.BoolToInt(bitSeq.GetBit(i))
		encodedDataAsNum += float64(bit) * math.Pow(2.0, float64(i+1)*-1.0)
	}

	var buffer bytes.Buffer
	end := 1.0
	start := 0.0
	for {
		width := end - start

		//encodedDataAsNum = (encodedDataAsNum - inter.lowerBound) / inter.width
		scaledData := (encodedDataAsNum - start) / width
		sym, inter := getSymbolWithinRange(scaledData, symToCumulativeProb)
		if sym == ENDSYMBOL {
			break
		}
		buffer.WriteByte(byte(sym))
		end = start + (width * inter.end)
		start = start + (width * inter.start)

	}

	return buffer.Bytes(), true
}

// Helpers
func getCumulativeProbabilitiesFromProbMap(probabilityMap *map[uint16]float64) map[uint16]probInterval {

	symToCumulativeProb := make(map[uint16]probInterval)
	prevEnd := 0.0
	sortedMapKeys := compressionutils.GetArrayOfSortedMapKeys(probabilityMap) // we need to guarantee order for because each interval needs to be divided up the same way for a static model

	for _, key := range sortedMapKeys {
		sym := uint16(key)
		symProb := (*probabilityMap)[sym]
		symToCumulativeProb[sym] = probInterval{prevEnd, prevEnd + symProb, symProb}
		prevEnd += symProb
	}

	//symToCumulativeProb[ENDSYMBOL] = probInterval{prevEnd, prevEnd + 1, 1}
	return symToCumulativeProb
}

func getSymbolWithinRange(num float64, symToCumulativeProb map[uint16]probInterval) (uint16, probInterval) {
	for sym, inter := range symToCumulativeProb {
		if num >= inter.start && num < inter.end {
			return sym, inter
		}
	}
	return 0, probInterval{}
}

func canCompressData(probabilityMap map[uint16]float64, srcDataLen float64) (float64, bool) {

	entropy := compressionutils.CalculateEntropyFromProbabilities(probabilityMap)
	//entropy := 0.0

	predictedCompressedSize := (srcDataLen * entropy) / float64(bitstructs.BYTE_LENGTH)
	fmt.Printf("Compression cannot do better than %v bits/symbol\n", entropy)
	fmt.Printf("Predicted compressed size (without probability table) = %v bytes\n", predictedCompressedSize)

	if int(entropy) >= bitstructs.BYTE_LENGTH {
		fmt.Printf("Cannot encode this as the compressed data length will probably be bigger than the original data since\n")
		fmt.Printf("Compression not effective for %v bits/symbol\n", entropy)
		return entropy, false
	}

	return entropy, true
}
