package core

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func EncodeBase64ToUrlString(input []byte) string {
	testing := string(input)
	fmt.Println(testing)

	result := base64.StdEncoding.EncodeToString(input)

	urlReplacer := strings.NewReplacer(
		"+", "-",
		"/", "_",
		"=", "",
	).Replace

	return urlReplacer(result)
}

func DecodeBase64FromUrlString(input []byte) ([]byte, error) {
	urlReplacer := strings.NewReplacer(
		"-", "+",
		"_", "/",
	).Replace

	inputString := urlReplacer(string(input))

	remainder := 4 - len(inputString)%4

	var spaceBuilder strings.Builder
	if remainder < 4 {
		for range remainder {
			spaceBuilder.WriteByte('=')
		}
	}

	spacedResult := string(inputString) + spaceBuilder.String()

	return base64.StdEncoding.DecodeString(spacedResult)
}
