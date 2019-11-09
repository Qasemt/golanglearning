package main

import (
	"errors"
	"fmt"
	b "github.com/qasemt/binance"
	h "github.com/qasemt/helper"
	st "github.com/qasemt/stockwork"
	"os"

	"time"
)

var appIniStr = "app init"

func init() {
	fmt.Printf(appIniStr + "\n")
	//err := h.SetProxy("socks5://127.0.0.1:9150", true)
	err := h.SetProxy("socks5://127.0.0.1:10028", true) // psiphon
	if err != nil {
		return
	}
}
func runCoin(asset string) error {
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO


	var begin time.Time
	var end time.Time
	end = (time.Now())
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO minute
	//f := time.Minute * -30
	//begin = (end.Add(f))

	now1 := time.Now()
	e1 := st.GetAssetCreateLastCandel(asset, now1,h.M15)
	if e1 != nil {
		return e1
	}
	e2 := st.GetAssetCreateLastCandel(asset, now1, h.H1)
	if e2 != nil {
		return e2
	}

	e3 := st.GetAssetCreateLastCandel(asset, now1, h.H2)
	if e3 != nil {
		return e3
	}
	e4 := st.GetAssetCreateLastCandel(asset, now1, h.H4)
	if e4 != nil {
		return e4
	}
	e5 := st.GetAssetCreateLastCandel(asset, now1, h.D1)
	if e5 != nil {
		return e5
	}

	//:::::::::::::::::::::::::::::::::::::::: CRYPTO HOUR
	begin = (end.AddDate(0, 0, -30))
	e6 := st.GetAsset("BTCUSDT", begin, end, h.M15, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdtm15.csv")
	if e6 != nil {
		return e6
	}

	begin = (end.AddDate(0, 0, -30))
	e7 := st.GetAsset("BTCUSDT", begin, end, h.H1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth1.csv")
	if e7 != nil {
		return e7
	}

	begin = (end.AddDate(0, 0, -30))
	e8 := st.GetAsset("BTCUSDT", begin, end, h.H2, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth2.csv")
	if e8 != nil {
		return e8
	}
	begin = (end.AddDate(0, 0, -100))
	e9 := st.GetAsset("BTCUSDT", begin, end, h.H4, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdth4.csv")
	if e9 != nil {
		return e9
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO DAILY
	begin = (end.AddDate(-2, 0, 0))
	e10 := st.GetAssetYear("BTCUSDT", begin, end, h.D1, "D:\\workspace\\stock\\data\\crypto\\new\\btcusdt_d.csv")
	if e10 != nil {
		return e10
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO 1 min
	return nil
}
func doingLoadCoinWithTime() (bool, error) {
	err := runCoin("BTCUSDT")

	if err != nil {
		return false, err
	}

	tick := time.Tick(10 * time.Minute)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-tick:

			err := runCoin("BTCUSDT")

			if err != nil {
				//		return false, err
				fmt.Println(err)
			}
			fmt.Println("-> wait 10 minute")
		}
	}

}
func binanceV2(){
	f:=b.MakeCacheHourly("BTCUSDT",h.M15,h.H1,time.Duration(time.Hour* 24*1), time.Now())
	fmt.Println(f)

	s := b.MakeCacheBase15M("BTCUSDT", time.Duration(time.Hour* 24*10), time.Now())

	fmt.Println(s)

}
func commands(a []string ) error {
	if len(a) ==2 &&  a[0]=="crypto" && a[1] =="BTCUSDT" {
		e := runCoin(a[1])
		if e != nil {
			return e
		}
	}else if len(a) ==2 &&  a[0]=="tehran"   {
		if len(a[1]) ==0{
			return  errors.New(fmt.Sprintln("path src is empty"))
		}
		if len(a[2]) ==0{
			return  errors.New(fmt.Sprintln("path dest is empty"))
		}
		if len(a[3]) !=0 && a[3]!= "adj"{
			return  errors.New(fmt.Sprintln("arg 3 not valid"))
		}

		if a[3]=="" {
		e:=	st.RUNStock(a[1], a[2], false)
		if e!=nil {
			return e
		}
		}else{
			e:=	st.RUNStock(a[1], a[2], true)
			if e!=nil{
				return e
			}
		}
	}else {
		s := fmt.Sprintf("Help args : [crypto] [BTCUSDT]\nHelp args : [tehran] [src dir path ] [dst dir path]")
		return errors.New(s)
	}
	return nil
}
func main() {
 e :=commands(os.Args[1:])
 
 if e!=nil{
 	fmt.Printf(e.Error())
	 return
 }
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO
	/*_, e := doingLoadCoinWithTime()
	if e != nil {
		fmt.Println("Failed :( please try again \n", e)
	}*/

	//argsWithoutProg := os.Args[1:]

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
	//	e := runCoin()

	fmt.Println("finished :)")

}
