package main

import (
	"fmt"

	"./modules/stockwork"
)

var appIniStr = "app init"

func init() {
	fmt.Printf(appIniStr + "\n")
}
func main() {

	//testFunctions.RunTest()
	//testFunctions.RunTestInterface()
	//	testFunctions.RunTestGoRoot()
	//testFunctions.RunTestMutex()
	//	testFunctions.RunTestCSV()

	//var path_dst_dir string = "D:/workspace/stock/tseclient/tmp/"
	stockwork.RUNStock("D:/workspace/stock/tseclient/normal/", "D:/out/", false)
	stockwork.RUNStock("D:/workspace/stock/tseclient/Adjusted/", "D:/out2/", true)
}
