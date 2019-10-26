package testFunctions

import (
	"fmt"
	"regexp"
)

func TestRegex() {
	matched, err := regexp.MatchString(`a.b`, "aaxbb")
	fmt.Println(matched) // true
	fmt.Println(err)     // nil (regexp is valid)
}
