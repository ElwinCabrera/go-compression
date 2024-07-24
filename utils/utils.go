package utils

import "math"

func GetSymbolFrequencyMap(data *[]byte) map[byte]int {
	var frequencyMap = make(map[byte]int)
	for _, sym := range *data {
		frequencyMap[sym]++
	}
	return frequencyMap
}

func GetSymbolProbabilityMap(data *[]byte) map[byte]float64 {
	var probabilityMap = make(map[byte]float64)
	freqMap := GetSymbolFrequencyMap(data)
	for sym, freq := range freqMap {
		probabilityMap[sym] = float64(freq) / float64(len(*data))
	}

	return probabilityMap
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
