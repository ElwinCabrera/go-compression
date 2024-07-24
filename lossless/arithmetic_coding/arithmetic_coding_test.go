package arithmeticcoding

import (
	"fmt"
	"testing"
)

func TestEncoding(t *testing.T) {
	data := []byte{'B', 'A', 'C', 'A', 0x00}

	symInfo0 := &SymbolInfo{'A', 0.4, 0.0}
	symInfo1 := &SymbolInfo{'B', 0.3, 0.0}
	symInfo2 := &SymbolInfo{'C', 0.2, 0.0}
	symInfo3 := &SymbolInfo{0x00, 0.1, 0.0}
	probabilityModel := []*SymbolInfo{symInfo0, symInfo1, symInfo2, symInfo3}

	res, canCompress := EncodeWithProbabilityModel(&data, probabilityModel)

	if canCompress {
		fmt.Printf("Result: %v\n", res)
	}

	decodedRes := DecodeWithProbabilityModel(res[0], probabilityModel)

	fmt.Printf("Result: %v\n", string(decodedRes))

}
