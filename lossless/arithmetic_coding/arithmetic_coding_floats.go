package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/utils"
	bitstructs "github.com/ElwinCabrera/go-data-structs/bit-structs"
	"math"
)

type Interval struct {
	lowerBound, upperBound, width float64
}

func Encode(srcData *[]byte) ([]byte, bool) {
	probabilityMap, _ := utils.GetSymbolProbabilityMap(srcData, true)
	return EncodeWithProbabilityModel(srcData, probabilityMap)
}

func Decode(encodedData *[]byte) []byte {
	//TODO: Need to find a way to send the probability map
	probabilityMap := map[byte]float64{}
	return DecodeWithProbabilityModel(encodedData, &probabilityMap, 0x00)
}

func EncodeWithProbabilityModel(srcData *[]byte, probabilityMap *map[byte]float64) ([]byte, bool) {
	_, canCompress := canCompressData(*probabilityMap, float64(len(*srcData)))
	if !canCompress {
		return []byte{}, false
	}
	symToCumProb := getCumulativeProbabilitiesFromProbMap(probabilityMap)
	return encode(srcData, symToCumProb)
}

func DecodeWithProbabilityModel(encodedBytes *[]byte, probabilityMap *map[byte]float64, endSymbol byte) []byte {
	cm := getCumulativeProbabilitiesFromProbMap(probabilityMap)
	res, _ := decode(encodedBytes, cm, endSymbol)
	return res
}

func encode(srcData *[]byte, symToCumulativeProb map[byte]Interval) ([]byte, bool) {

	upper := 1.0
	lower := 0.0

	for _, sym := range *srcData {
		width := upper - lower
		upper = lower + (width * symToCumulativeProb[sym].upperBound)
		lower = lower + (width * symToCumulativeProb[sym].lowerBound)

		if lower == upper {
			fmt.Printf("Reached maximum number of symbols that can be encoded using 64-bit floating point math\n")
		}
	}
	fmt.Printf("lowbound: %v, upperbound: %v, width: %v\n", lower, upper, upper-lower)
	//encodedData := utils.GetMidPointProtectingAgainstOverflow(lower, upper)

	bitSeq := bitstructs.NewDynamicBitSequence()
	for {
		if lower > 0.5 && upper > 0.5 {
			bitSeq.PushBitEnd(1)
			lower = (lower / 0.5) - 1
			upper = (upper / 0.5) - 1
		} else if lower < 0.5 && upper < 0.5 {
			bitSeq.PushBitEnd(0)
			lower = lower / 0.5
			upper = upper / 0.5
		} else if lower < 0.5 && upper > 0.5 {
			lower = lower / 0.5
			upper = (upper / 0.5) - 1
		} else {
			break
		}

	}
	if lower > 0.5 {
		bitSeq.PushBitEnd(1)
	} else {
		bitSeq.PushBitEnd(0)
	}
	fmt.Printf("Binary representation: 0.%v\n", bitSeq)

	paddingBits := 0
	if bitSeq.GetNumBits()%bitstructs.BYTE_LENGTH != 0 {
		paddingBits = bitstructs.BYTE_LENGTH - (bitSeq.GetNumBits() % bitstructs.BYTE_LENGTH)
		bitSeq.ExpandNumOfBitsToUseRemainingBitsInLastByte()
		for i := paddingBits; i > 0; i-- {
			bitSeq.ShiftLeft()
		}
	}

	setBit := true
	if lower > 0.5 {
		setBit = false
	}
	for i := 0; i < paddingBits; i++ {
		bitSeq.SetBit(i, setBit)
	}

	fmt.Printf("Binary representation (after adding padding): 0.%v\n", bitSeq)

	return bitSeq.GetBitSeq(), true
}

func decode(encodedBytes *[]byte, symToCumulativeProb map[byte]Interval, endSymbol byte) ([]byte, bool) {

	encodedDataAsNum := 0.0
	bitLen := len(*encodedBytes) * bitstructs.BYTE_LENGTH
	bitSeq := bitstructs.NewBitSequenceFromByteArray(encodedBytes, bitLen)
	for i := bitLen - 1; i >= 0; i-- {
		bit := bitstructs.BoolToInt(bitSeq.GetBit(i))
		encodedDataAsNum += float64(bit) * math.Pow(2.0, float64(i-bitLen))
	}

	var buffer bytes.Buffer
	for {
		sym, inter := getSymbolWithinRange(encodedDataAsNum, symToCumulativeProb)
		encodedDataAsNum = (encodedDataAsNum - inter.lowerBound) / inter.width
		if sym == endSymbol {
			break
		}
		buffer.WriteByte(sym)
	}

	return buffer.Bytes(), true
}

// Helpers
func getCumulativeProbabilitiesFromProbMap(probabilityMap *map[byte]float64) map[byte]Interval {

	symToCumulativeProb := make(map[byte]Interval)
	prevEnd := 0.0
	sortedMapKeys := utils.GetArrayOfSortedMapKeys(probabilityMap) // we need to guarantee order for because each interval needs to be divided up the same way for a static model
	for _, key := range sortedMapKeys {
		symProb := (*probabilityMap)[byte(key)]
		symToCumulativeProb[byte(key)] = Interval{prevEnd, prevEnd + symProb, symProb}
		prevEnd += symProb
	}
	return symToCumulativeProb
}

func getSymbolWithinRange(num float64, symToCumulativeProb map[byte]Interval) (byte, Interval) {
	for sym, inter := range symToCumulativeProb {
		if num >= inter.lowerBound && num < inter.upperBound {
			return sym, inter
		}
	}
	return 0, Interval{}
}

func canCompressData(probabilityMap map[byte]float64, srcDataLen float64) (float64, bool) {

	entropy := utils.CalculateEntropyFromProbabilities(probabilityMap)
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
