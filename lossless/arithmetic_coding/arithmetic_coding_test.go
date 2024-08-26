package arithmeticcoding

import (
	"bytes"
	"fmt"
	"github.com/ElwinCabrera/go-compression/compressionutils"
	testingutils "github.com/ElwinCabrera/go-compression/testing_utils"
	"maps"
	"testing"
)

func testSerializationAndDeserialization(t *testing.T, testingData *[]byte) {
	freqMap := compressionutils.GetSymbolFrequencyMap(testingData)

	serializedFreqTable := serializeFrequencyTable(freqMap)

	deSerializedFreqTable, serializedLen := deserializeFrequencyTable(&serializedFreqTable)

	if serializedLen != len(serializedFreqTable) {
		t.Fatalf("When Deserializing the expected total original serialized length does not equal to the original serialized lenth. Got length %v but expected %v. Expected table: %v ... but Got table: %v\n", serializedLen, len(serializedFreqTable), freqMap, deSerializedFreqTable)

	}
	if !maps.Equal(*freqMap, deSerializedFreqTable) {
		t.Fatalf("Serialization and Deserialization of frequency tables are not equal. Expected %v, got %v\n", freqMap, deSerializedFreqTable)
	}
}

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
		testSerializationAndDeserialization(t, &data)
		fmt.Printf("Serialization and Deserialization of table test PASS for dataset #%v with size of %v+1 bytes\n", i, len(data))
		testCompressAndDecompress(t, &data)
		fmt.Printf("Compression and Decompression test PASS for dataset #%v with size of %v+1 bytes\n\n", i, len(data))
	}
}
