package compressionutils

import (
	"math"
	"sort"
)

type FractionParts struct {
	numerator   uint64
	denominator uint64
}

func GetSymbolFrequencyMap(data *[]byte) *map[uint16]uint64 {
	var frequencyMap = make(map[uint16]uint64)
	for _, sym := range *data {
		frequencyMap[uint16(sym)]++
	}
	return &frequencyMap
}

func GetSymbolProbMapFromFreqMap(frequencyMap *map[uint16]uint64, dataLen int) *map[uint16]float64 {
	var probabilityMap = make(map[uint16]float64)
	for sym, freq := range *frequencyMap {
		probabilityMap[sym] = float64(freq) / float64(dataLen)
	}
	return &probabilityMap
}

//func GetSymbolProbabilityMapFromData(data *[]byte) *map[byte]float64 {
//	freqMap := GetSymbolFrequencyMap(data)
//	return GetSymbolProbMapFromFreqMap(freqMap, len(*data))
//}

func CalculateEntropyFromProbabilities(probabilitiesMap map[uint16]float64) float64 {
	var entropy = 0.0

	for _, prob := range probabilitiesMap {
		entropy += prob * math.Log2(1/prob)
	}
	return entropy
}

func GetMidPointProtectingAgainstOverflow(start, end float64) float64 {
	return ((end - start) / 2) + start
}

func FindAndGetEndSymbol[T any](symbolMap *map[uint16]T) uint16 {
	endSymbol := uint16(0)

	_, is0InMap := (*symbolMap)[0]
	_, is255InMap := (*symbolMap)[0xFF]

	if !is0InMap {
		endSymbol = 0x00
	} else if !is255InMap {
		endSymbol = 0xFF
	} else {

		for ; endSymbol < 255; endSymbol++ {
			_, isInMap := (*symbolMap)[endSymbol]
			if !isInMap {
				break
			}
		}
	}
	return endSymbol
}

func GetArrayOfSortedMapKeys[V any](mapToSortKey *map[uint16]V) []int {
	keys := make([]int, 0)
	for k, _ := range *mapToSortKey {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	//sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys
}

func SimplifyFraction(numerator, denominator uint64) (uint64, uint64) {
	gcf := findGCF([]uint64{numerator, denominator})
	return numerator / gcf, denominator / gcf
}

func findGCF(nums []uint64) uint64 {
	loopEnd := uint64(0)
	for _, num := range nums {
		if loopEnd < num {
			loopEnd = num
		}
	}
	for i := loopEnd; i > 0; i-- {
		foundGCF := true
		for _, num := range nums {
			if num%i != 0 {
				foundGCF = false
				break
			}
		}
		if foundGCF {
			return i
		}
	}
	return 1

}

func findLowestSharedMultiple(nums []uint64, maxIter int) int {
	lcm := 1
	for i := 2; i <= maxIter; i++ {
		isCommonMultiple := true
		for _, num := range nums {
			if num%uint64(i) != 0 {
				isCommonMultiple = false
				break
			}
		}
		if isCommonMultiple {
			lcm = i
			break
		}
	}
	return lcm
}
