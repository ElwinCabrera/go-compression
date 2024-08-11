package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/compressionutils"
	testingutils "github.com/ElwinCabrera/go-compression/testing_utils"
	"testing"
)

func testEncodeAndDecode(t *testing.T, testingData *[]byte) {
	freqMap := compressionutils.GetSymbolFrequencyMap(testingData)

	res, canCompress := EncodeWithProbabilityModel(testingData, freqMap, true)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	//fmt.Printf("Encoded result: %v\n", res)

	decodedRes, _ := DecodeWithProbabilityModel(&res, freqMap, uint(len(*testingData)))

	if !bytes.Equal(*testingData, decodedRes) {
		t.Fatalf("Decoded data does not match original data.\n") // Got %v, expected %v\n", string(decodedRes), string(*testingData))
	}
}

func testCompressAndDecompress(t *testing.T, testingData *[]byte) {

	res, canCompress := Compress(testingData)
	if !canCompress {
		t.Fatalf("Can't compress data using arithmetic coding.")
	}

	//fmt.Printf("Encoded result: %v\n", res)

	decodedRes, _ := Decompress(&res)

	if !bytes.Equal(*testingData, decodedRes) {
		t.Fatalf("Decompressed data does not match original data.\n") // Got %v, expected %v\n", string(decodedRes), string(*testingData))
	}
}

func TestArithmeticCoding(t *testing.T) {

	testingData := testingutils.GetSomeTestData()
	for i, data := range testingData {
		//testEncodeAndDecode(t, &data)
		testCompressAndDecompress(t, &data)
		fmt.Printf("Test %v Compress and Decompress of %v bytes pass\n\n", i, len(data)+1)
	}
}
