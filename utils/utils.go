package utils

import (
	"math"
	"sort"
)

func GetSymbolFrequencyMap(data *[]byte, appendEndSymbol bool) (*map[byte]int, byte) {
	var frequencyMap = make(map[byte]int)
	for _, sym := range *data {
		frequencyMap[sym]++
	}

	endSymbol := byte(0)
	if appendEndSymbol {
		endSymbol = findEndSymbol(&frequencyMap)
		frequencyMap[endSymbol] = 1
	}
	return &frequencyMap, endSymbol //returning as a pointer to preserve map order since go does not have a ordered map type like in c++
}

func GetSymbolProbabilityMap(data *[]byte, appendEndSymbol bool) (*map[byte]float64, byte) {
	dataLen := float64(len(*data))
	if appendEndSymbol {
		dataLen++
	}
	var probabilityMap = make(map[byte]float64)
	freqMap, endSymbol := GetSymbolFrequencyMap(data, appendEndSymbol)

	for sym, freq := range *freqMap {
		probabilityMap[sym] = float64(freq) / dataLen
	}
	return &probabilityMap, endSymbol
}

func CalculateEntropyFromProbabilities(probabilitiesMap map[byte]float64) float64 {
	var entropy = 0.0

	for _, prob := range probabilitiesMap {
		entropy += prob * math.Log2(1/prob)
	}
	return entropy
}

func GetMidPointProtectingAgainstOverflow(start, end float64) float64 {
	return ((end - start) / 2) + start
}

func findEndSymbol[T any](symbolMap *map[byte]T) byte {
	endSymbol := byte(0)

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

func GetArrayOfSortedMapKeys[T any](mapToSortKey *map[byte]T) []int {
	keys := make([]int, 0)
	for k, _ := range *mapToSortKey {
		keys = append(keys, int(k))
	}
	//sort.Ints(keys)
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys
}
