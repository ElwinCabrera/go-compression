package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/compressionutils"
	"testing"
)

func test0(t *testing.T) {
	data := []byte{'A', 'B', 'A', 'B', 'A', 'C'}
	freqMap := compressionutils.GetSymbolFrequencyMap(&data)
	(*freqMap)[ENDSYMBOL] = 1
	probMap := compressionutils.GetSymbolProbMapFromFreqMap(freqMap, len(data)+1)

	//Range should be [0.21154, 0.21167) if symbols ar sorted
	////bit sequence should be 0.00110110001 01...... if symbols ar sorted
	res, canCompress := finiteEncodeWithProbabilityModel(&data, probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := finiteDecodeWithProbabilityModel(&res, probMap)

	if !bytes.Equal(data, decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Succesfully Decoded: %v\n", string(decodedRes))
}

func test1(t *testing.T) {
	data := []byte{'A', ' ', 'S', 'A', 'D', ' ', 'S', 'A', 'L', 'A', 'D'}
	freqMap := compressionutils.GetSymbolFrequencyMap(&data)
	(*freqMap)[ENDSYMBOL] = 1
	probMap := compressionutils.GetSymbolProbMapFromFreqMap(freqMap, len(data)+1)
	//Range should be [, ) if symbols ar sorted
	////bit sequence should be 0.100110110001 01 ...... if symbols ar sorted

	res, canCompress := finiteEncodeWithProbabilityModel(&data, probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := finiteDecodeWithProbabilityModel(&res, probMap)

	if !bytes.Equal(data, decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func test2(t *testing.T) {
	data := []byte{'2', '1', 0x00}
	probMap := map[uint16]float64{'1': 0.4, '2': 0.4, ENDSYMBOL: 0.2}

	//Range should be [.68 , .712)
	//bit sequence should be 0.101100 .....

	res, canCompress := finiteEncodeWithProbabilityModel(&data, &probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := finiteDecodeWithProbabilityModel(&res, &probMap)

	if !bytes.Equal(data, decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func test3(t *testing.T) {
	data := []byte{'B', 'A', 'C', 'A'}
	//data := []byte{'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 0x00}
	probabilityMap := map[uint16]float64{'A': 0.4, 'B': 0.3, 'C': 0.2, ENDSYMBOL: 0.1} //range should be

	res, canCompress := finiteEncodeWithProbabilityModel(&data, &probabilityMap)

	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := finiteDecodeWithProbabilityModel(&res, &probabilityMap)

	if !bytes.Equal(data, decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func TestEncodingDecoding(t *testing.T) {
	//test0(t)
	//test1(t)
	test2(t)
	test3(t)
}
