package huffman

import (
	"bytes"
	"fmt"
	testinguutils "github.com/ElwinCabrera/go-compression/testing_utils"
	"github.com/ElwinCabrera/go-compression/utils"
	"math"

	"testing"

	"github.com/ElwinCabrera/go-data-structs/trees"
)

func testCompressAndDecompress(t *testing.T, testData *[]byte) {

	compressedData, canCompress := Compress(testData)

	unCompressedData := Decompress(&compressedData)

	if !bytes.Equal(*testData, *unCompressedData) {
		t.Errorf("Decompress failed. Expected %v, got %v", *testData, *unCompressedData)
	}

	fmt.Printf("Compressed size: %d bytes\n", len(compressedData))
	if len(compressedData) >= len(*testData) || !canCompress {
		t.Logf("Note: compressed data is larger than the uncompressed data %v bytes compressed vs. %v bytes uncompressed\n", len(compressedData), len(*testData))
	}
}

func testSerializeDeSerializeOfHuffmanCodes(t *testing.T, testData *[]byte) {

	//Get the frequency each byte appears in the data we want to compress
	freqMap, _ := utils.GetSymbolFrequencyMap(testData, false)

	ht := trees.NewHuffmanTreeFromFrequencyMap(*freqMap)
	huffmanCodes := ht.GetHuffmanCodes()

	serializedHuffmanCodes := serializeHuffmanCodes(huffmanCodes)

	deserializedHuffmanCodes, serializedDataLen := deSerializeHuffmanCodesFromByteArray(&serializedHuffmanCodes)

	if serializedDataLen != len(serializedHuffmanCodes) {
		t.Fatalf("byte array length of serialized huffmancode does not match the one got when deserializing. got %v expected %v ", serializedDataLen, len(serializedHuffmanCodes))
	}

	for k, v1 := range huffmanCodes {
		v2, ok := deserializedHuffmanCodes[k]
		if !ok {
			t.Fatalf("compression code for key %v not found in deserialized compression codes", k)
		}
		if v1.GetXBytes(8) != v2.GetXBytes(8) {
			t.Fatalf("Deserialized compression codes are not the dame as the original. expected %v but got %v when deserialized for key %v\n", v1, v2, k)
		}
	}

	//Below are just some stats we can calculate just for fun

	// Lets calculate entropy (minimum number of bits needed per symbol)
	// First we calculate the maximum entropy and then the actual entropy of the system

	//This is the maximum entropy we can expect
	//Minimum num of bits needed per symbol symbols when there is uniform probability i.e all symbols have same probability or each symbol is equally likely to occur 1/<num of unique symbols found>
	probabilityOfEachSymbolOccurring := 1 / float64(len(*freqMap))
	inverseP := 1 / probabilityOfEachSymbolOccurring
	maxEntropyWithUniformProb := probabilityOfEachSymbolOccurring * math.Log2(inverseP) * float64(len(*freqMap))

	//Now let's calculate the actual entropy
	actualEntropy := 0.0
	for _, freq := range *freqMap {
		probabilityOfSymbolOccurring := float64(freq) / float64(len(*testData))
		inverseP = 1 / probabilityOfSymbolOccurring
		actualEntropy += probabilityOfSymbolOccurring * math.Log2(inverseP)
	}

	averageBitLengthPerSym := 0.0
	for sym, bs := range huffmanCodes {
		symWeight := float64((*freqMap)[sym]) / float64(len(*testData))
		averageBitLengthPerSym += symWeight * float64(bs.GetNumBits())
	}
	//averageBitLength needs to be >= entropy and
	//entropy needs to be <= max entropy
	//Otherwise, it's impossible
	if actualEntropy > maxEntropyWithUniformProb || averageBitLengthPerSym < actualEntropy {
		t.Logf("Impossible")
	}

	//anything less than the entropy value is impossible

	fmt.Printf("\n\tData size is %v containing %v unique symbols and each symbol can be any number between 0-%v\n", len(*testData), len(*freqMap), len(*freqMap)-1)
	fmt.Printf("\tWorst case: Max Entropy with (uniform probability for each symbol): %f bits/sym\n", maxEntropyWithUniformProb)
	fmt.Printf("\tActual entropy (actual min num of bits/sym) :%f bits/sym\n", actualEntropy)
	fmt.Printf("\tAverage bits/sym: %f\n", averageBitLengthPerSym)

	predictedCompressedDataByteLen := (actualEntropy * float64(len(*testData))) / 8
	if actualEntropy == 0 {
		predictedCompressedDataByteLen = float64(len(*testData)) / 8
	}
	fmt.Printf("\tPredicted compressed size of just data : %f bytes\n", predictedCompressedDataByteLen)
	compressedData, _ := Compress(testData)
	compressedDataSize := len(compressedData) - len(serializedHuffmanCodes) - 1 // minus one because we add and extra byte at the end
	fmt.Printf("\tActual compressed size (not including serialized compression codes and extra 1 byte): %v bytes\n", compressedDataSize)

	percentError := ((predictedCompressedDataByteLen - float64(compressedDataSize)) / float64(compressedDataSize)) * 100
	percentError = math.Abs(percentError)
	fmt.Printf("\tPercent error in size: %v percent\n", percentError)
	fmt.Printf("\tEfficiency %f percent\n", (actualEntropy/averageBitLengthPerSym)*100)

	fmt.Printf("\tSerialized compression code size: %v bytes\n", len(serializedHuffmanCodes))
	fmt.Printf("\tTotal compressed size: %v bytes\n", len(compressedData))

}

func TestAll(t *testing.T) {

	for i, data := range testinguutils.GetSomeTestData() {
		fmt.Printf("\nTesting data at index %v ... ", i)
		testSerializeDeSerializeOfHuffmanCodes(t, &data)
		testCompressAndDecompress(t, &data)
		fmt.Println("Test Pass")
	}

}
