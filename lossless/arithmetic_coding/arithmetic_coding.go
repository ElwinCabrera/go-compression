package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/utils"
	bitstructs "github.com/ElwinCabrera/go-data-structs/bit-structs"
	"github.com/ElwinCabrera/go-data-structs/list"
)

type SymbolInfo struct {
	Symbol        byte
	Probability   float64
	intervalStart float64
}

func Encode(srcData *[]byte) ([]byte, bool) {
	probabilityMap := utils.GetSymbolProbabilityMap(srcData)
	addEndSymbolToProbabilityMap(&probabilityMap)
	dll := list.NewDoublyLinkedList()
	for sym, prob := range probabilityMap {
		dummyVal := 0.0
		symbolInfo := &SymbolInfo{sym, prob, dummyVal}
		dll.InsertSortedDescBasedOnNodeWeight(symbolInfo, prob)
	}
	//return EncodeWithProbabilityModel(srcData, probabilityMap)
	return []byte{}, false
}

func Decode(encodedData *[]byte) []byte {

	return []byte{}
}

func EncodeWithProbabilityModel(srcData *[]byte, probabilityModel []*SymbolInfo) ([]float64, bool) {

	_, canCompress := canCompressData(probabilityModel, float64(len(*srcData)))
	if !canCompress {
		return []float64{}, false
	}

	updateSymbolIntervalStart(&probabilityModel, 0.0, 1.0)

	symbolToSymbolInfo := make(map[byte]*SymbolInfo)
	for _, symbolInfo := range probabilityModel {
		symbolToSymbolInfo[symbolInfo.Symbol] = symbolInfo
	}

	start := 0.0
	end := 1.0
	for _, sym := range *srcData {
		width := end - start
		start = symbolToSymbolInfo[sym].intervalStart
		end = start + (symbolToSymbolInfo[sym].Probability * width)
		updateSymbolIntervalStart(&probabilityModel, start, end-start)
	}

	encodedData := utils.GetMidPointProtectingAgainstOverflow(start, end)
	return []float64{encodedData}, true
	//return []float64{start, end}, true
}

func DecodeWithProbabilityModel(encodedData float64, probabilityModel []*SymbolInfo) []byte {
	updateSymbolIntervalStart(&probabilityModel, 0.0, 1.0)

	var buffer bytes.Buffer

	for {
		symInfo := getSymbolWithinRange(encodedData, &probabilityModel)
		if symInfo.Symbol == 0x00 {
			break
		}
		buffer.WriteByte(symInfo.Symbol)
		encodedData = (encodedData - symInfo.intervalStart) / symInfo.Probability

	}
	//return []byte{}
	return buffer.Bytes()
}

// helpers
func updateSymbolIntervalStart(probabilityModel *[]*SymbolInfo, newStart, newWidth float64) {
	prevEnd := newStart
	for _, symInfo := range *probabilityModel {
		symInfo.intervalStart = prevEnd
		prevEnd += symInfo.Probability * newWidth
	}
}

func getSymbolWithinRange(num float64, probabilityModel *[]*SymbolInfo) *SymbolInfo {
	for _, symInfo := range *probabilityModel {
		start := symInfo.intervalStart
		end := start + symInfo.Probability
		if num >= start && num < end {
			return symInfo
		}
	}
	return nil
}

func addEndSymbolToProbabilityMap(probabilityMap *map[uint8]float64) {
	newSymbolProbability := 1 / float64(len(*probabilityMap))
	_, is0InMap := (*probabilityMap)[0]
	_, is255InMap := (*probabilityMap)[0xFF]

	if !is0InMap {
		(*probabilityMap)[0] = newSymbolProbability
	} else if !is255InMap {
		(*probabilityMap)[0xFF] = newSymbolProbability
	} else {
		//TODO: find any symbol that is not in probability map
	}
}

func canCompressData(probabilityModel []*SymbolInfo, srcDataLen float64) (float64, bool) {

	//entropy := utils.CalculateEntropyFromProbabilities(probabilityModel)
	entropy := 0.0

	predictedCompressedSize := (srcDataLen * entropy) / float64(bitstructs.BYTE_LENGTH)
	fmt.Printf("Compression cannot do better than %v bits/symbol\n", entropy)
	fmt.Printf("Pridicted compressed size (without probability table) = %v bytes\n", predictedCompressedSize)

	if int(entropy) >= bitstructs.BYTE_LENGTH {
		fmt.Printf("Cannot encode this as the compressed data length will probably be bigger than the original data since\n")
		fmt.Printf("Compression not effective for %v bits/symbol\n", entropy)
		return entropy, false
	}

	return entropy, true
}
