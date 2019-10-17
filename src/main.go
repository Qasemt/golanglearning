package main

import (
	"fmt"

	"./myfunctions"
)

var appIniStr = "app init"

func init() {
	fmt.Printf(appIniStr + "\n")
}
func main() {

	myfunctions.RunTest()

}
