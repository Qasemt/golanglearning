package main

import (
	"fmt"
	"time"

	h "./pk/helper"
	st "./pk/stockwork"
	//"./testFunctions"
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
	//stockwork.RUNStock("D:/workspace/stock/tseclient/normal/", "D:/out/", false)
	//stockwork.RUNStock("D:/workspace/stock/tseclient/Adjusted/", "D:/out2/", true)
	//testFunctions.TestRegex()

	//:::::::::::::::::::::::::::::::::::::::: FOREX
	//stockwork.ConvertStoockTODT7("D:/workspace/stock/data/forex/ff.csv", "D:/workspace/stock/data/forex/ff11.csv")

	//:::::::::::::::::::::::::::::::::::::::: CRYPTO
	h.SetProxy("http://191.102.106.1:8181")
	var start time.Time
	end := (time.Now())
	start = (time.Now().AddDate(0, 0, -100))
	//stockwork.GetAsset("BTCUSDT", start_num, end, "4h", "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_h4.csv")
	//	stockwork.GetAsset("BTCUSDT", start_num, end, "1h", "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_h1.csv")
	//	stockwork.GetAsset("BTCUSDT", start_num, end, "2h", "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_h2.csv")
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO DAILY
	start = (time.Now().AddDate(-2, 0, 0))
	st.GetAssetYear("BTCUSDT", start, end, "1d", "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_d.csv")

}
