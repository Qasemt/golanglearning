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

	variable := myfunctions.SayHello()
	fmt.Println(variable)

	myfunctions.TestSwitch()
	myfunctions.Testarray()

}
