package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/utils"
	"testing"
)

func test1(t *testing.T) {
	data := []byte{'A', ' ', 'S', 'A', 'D', ' ', 'S', 'A', 'L', 'A', 'D'}
	probMap, endSymbol := utils.GetSymbolProbabilityMap(&data, true)
	data = append(data, endSymbol)
	//Range should be [0.21154, 0.21167)
	////bit sequence should be 0.100110110001 01 ......
	res, canCompress := EncodeWithProbabilityModel(&data, probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := DecodeWithProbabilityModel(&res, probMap, endSymbol)

	if !bytes.Equal(data[:len(data)-1], decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func test0(t *testing.T) {
	data := []byte{'A', 'B', 'A', 'B', 'A', 'C'}
	probMap, endSymbol := utils.GetSymbolProbabilityMap(&data, true)
	data = append(data, endSymbol)

	//Range should be [0.21154, 0.21167)
	////bit sequence should be 0.00110110001 01......
	res, canCompress := EncodeWithProbabilityModel(&data, probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := DecodeWithProbabilityModel(&res, probMap, endSymbol)

	if !bytes.Equal(data[:len(data)-1], decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Succesfully Decoded: %v\n", string(decodedRes))
}

func test2(t *testing.T) {
	data := []byte{'2', '1', 0x00}
	probMap := map[byte]float64{'1': 0.4, '2': 0.4, 0x00: 0.2}

	//Range should be [.68 , .712)
	//bit sequence should be 0.101100 .....

	res, canCompress := EncodeWithProbabilityModel(&data, &probMap)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := DecodeWithProbabilityModel(&res, &probMap, 0x00)

	if !bytes.Equal(data[:len(data)-1], decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func test3(t *testing.T) {
	//data := []byte{'B', 'A', 'C', 'A', 0x00}
	data := []byte{'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 'C', 'A', 'B', 'A', 0x00}
	probabilityMap := map[byte]float64{'A': 0.4, 'B': 0.3, 'C': 0.2, 0x00: 0.1} //range should be

	res, canCompress := EncodeWithProbabilityModel(&data, &probabilityMap)

	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	fmt.Printf("Encoded result: %v\n", res)

	decodedRes := DecodeWithProbabilityModel(&res, &probabilityMap, 0x00)

	if !bytes.Equal(data[:len(data)-1], decodedRes) {
		t.Fatalf("Decoded data does not match original data. Got %v, expected %v\n", string(decodedRes), string(data))
	}

	fmt.Printf("Decoded result: %v\n", string(decodedRes))

}

func TestEncodingDecoding(t *testing.T) {
	test3(t)
}
