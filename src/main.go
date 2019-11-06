package main

import (
	"fmt"
	"time"

	h "./pk/helper"
	st "./pk/stockwork"
)

var appIniStr = "app init"

func init() {
	fmt.Printf(appIniStr + "\n")
}
func runCoin() {
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO
	h.SetProxy("http://191.102.106.1:8181")
	var begin time.Time
	var end time.Time
	end = (time.Now())
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO minute
	//f := time.Minute * -30
	//begin = (end.Add(f))
	fmt.Println(begin, end)
	now1 := time.Now()
	st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H1)
	st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H2)
	st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H4)
	st.GetAssetCreateLastCandel("BTCUSDT", now1, st.D1)

	//:::::::::::::::::::::::::::::::::::::::: CRYPTO HOUR
	begin = (time.Now().AddDate(0, 0, -100))
	st.GetAsset("BTCUSDT", begin, end, st.H4, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth4.csv")
	begin = (time.Now().AddDate(0, 0, -30))
	st.GetAsset("BTCUSDT", begin, end, st.H1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth1.csv")
	begin = (time.Now().AddDate(0, 0, -30))
	st.GetAsset("BTCUSDT", begin, end, st.H2, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth2.csv")
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO DAILY
	begin = (time.Now().AddDate(-2, 0, 0))
	st.GetAssetYear("BTCUSDT", begin, end, st.D1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_d.csv")
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO 1 min
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
	runCoin()
}
