package format

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func PrintSectionHeader(input, symbol string){

	line := strings.Repeat(symbol, utf8.RuneCountInString(input) + 4)

	fmt.Println(line)
	fmt.Printf("%s %s %s\n", symbol, input, symbol)
	fmt.Println(line)
	fmt.Println()
}
