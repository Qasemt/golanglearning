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
func runCoin() error {
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO
	err := h.SetProxy("socks5://127.0.0.1:9150", true)
	if err != nil {
		return err
	}
	var begin time.Time
	var end time.Time
	end = (time.Now())
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO minute
	//f := time.Minute * -30
	//begin = (end.Add(f))

	now1 := time.Now()
	e1 := st.GetAssetCreateLastCandel("BTCUSDT", now1, st.M15)
	if e1 != nil {
		return e1
	}
	e2 := st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H1)
	if e2 != nil {
		return e2
	}
	e3 := st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H2)
	if e3 != nil {
		return e3
	}
	e4 := st.GetAssetCreateLastCandel("BTCUSDT", now1, st.H4)
	if e4 != nil {
		return e4
	}
	e5 := st.GetAssetCreateLastCandel("BTCUSDT", now1, st.D1)
	if e5 != nil {
		return e5
	}

	//:::::::::::::::::::::::::::::::::::::::: CRYPTO HOUR
	begin = (end.AddDate(0, 0, -30))
	e6 := st.GetAsset("BTCUSDT", begin, end, st.M15, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdtm15.csv")
	if e6 != nil {
		return e6
	}

	begin = (end.AddDate(0, 0, -30))
	e7 := st.GetAsset("BTCUSDT", begin, end, st.H1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth1.csv")
	if e7 != nil {
		return e7
	}

	begin = (end.AddDate(0, 0, -30))
	e8 := st.GetAsset("BTCUSDT", begin, end, st.H2, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth2.csv")
	if e8 != nil {
		return e8
	}
	begin = (end.AddDate(0, 0, -100))
	e9 := st.GetAsset("BTCUSDT", begin, end, st.H4, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth4.csv")
	if e9 != nil {
		return e9
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO DAILY
	begin = (end.AddDate(-2, 0, 0))
	e10 := st.GetAssetYear("BTCUSDT", begin, end, st.D1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_d.csv")
	if e10 != nil {
		return e10
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO 1 min
	return nil
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
	e := runCoin()
	if e != nil {
		fmt.Println("Failed :( please try again \n", e)
	}
}
