package main

import (
	"errors"
	"fmt"
	av "github.com/qasemt/avardstock"
	h "github.com/qasemt/helper"
	st "github.com/qasemt/stockwork"
	"os"
	"strings"
	"sync"
	"time"
)

var appIniStr = "app init"

func init() {
	fmt.Printf(appIniStr + "\n")
	//err := h.SetProxy("127.0.0.1:9150", true)
	//err := h.SetProxy("https://127.0.0.1:5051", false) // psiphon
	/*	if err != nil {
		return
	}*/
}

/*func doingLoadCoinWithTime() (bool, error) {
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

}*/
func testFunction() {
	//testFunctions.RunTest()
	//testFunctions.RunTestInterface()
	//	testFunctions.RunTestGoRoot()
	//testFunctions.RunTestMutex()
	//	testFunctions.RunTestCSV()

	//testFunctions.TestRegex()

}

//:::::::::::::::::::::::::::::::::::::::: FOREX
func forex() {
	//stockwork.ConvertStoockTODT7("D:/workspace/stock/data/forex/ff.csv", "D:/workspace/stock/data/forex/ff11.csv")
}

//:::::::::::::::::::::::::::::::::::::::: CONVERT STOCK TEHRAN FROM TSE
func tehranTSEC() {
	//var path_dst_dir string = "D:/workspace/stock/tseclient/tmp/"
	//stockwork.RUNStock("D:/workspace/stock/tseclient/normal/", "D:/out/", false)
	//stockwork.RUNStock("D:/workspace/stock/tseclient/Adjusted/", "D:/out2/", true)
}
func avardAssetProcess(parentWaitGroup *sync.WaitGroup, readfromLast bool, assetCode string, nameEn string, isIndex bool) error {
	//var id string ="IRO1GDIR0001"
	if nameEn == "" || assetCode == "" {
		parentWaitGroup.Done()
		return errors.New("field is empty ")

	}
	var databaseLock sync.Mutex
	var wg sync.WaitGroup
	if isIndex == true {
		wg.Add(2)
	} else {
		wg.Add(8)
	}

	go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*250), time.Now(), h.H1, h.Normal)
	go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*4000), time.Now(), h.D1, h.Normal)

	if isIndex == false {
		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*250), time.Now(), h.H2, h.Normal)
		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*360), time.Now(), h.H4, h.Normal)

		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*250), time.Now(), h.H1, h.Adj)
		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*250), time.Now(), h.H2, h.Adj)
		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*360), time.Now(), h.H4, h.Adj)
		go av.Make(&wg, &databaseLock, readfromLast, assetCode, nameEn, isIndex, -time.Duration(time.Hour*24*4000), time.Now(), h.D1, h.Adj)
	}
	wg.Wait()
	parentWaitGroup.Done()
	return nil
}
func avardMainProcess(readfromLast bool) error {

	list, e := av.ReadJsonWatchList()

	if e != nil {
		return errors.New(fmt.Sprintf("config not found "))
	}
	var wg sync.WaitGroup
	wg.Add(len(list.Tehran))

	for _, g := range list.Tehran {
		go avardAssetProcess(&wg, readfromLast, g.AssetCode, g.NameEn, g.IsIndex)
		/*	if e != nil {
			return e
		}*/
	}
	wg.Wait()
	return nil
}
func readArgs(a []string, key string) (string, bool) {
	for i := 1; i < len(a); i++ {
		if strings.HasPrefix(strings.ToLower(a[i]), key) {
			var p string
			if strings.Contains(strings.ToLower(a[i]), "=") {
				p = strings.Split(a[i], "=")[1]
				p = strings.Trim(p, `"`)
			}
			return p, true
			break
		}
	}

	return "", false
}

//___________________________________________________________________
func commands(a []string) error {
	if len(a) > 0 && strings.ToLower(a[0]) == "crypto" {

		if v, ok := readArgs(a, "cachepath="); ok {
			h.SetRootCache(v)
		}
		//proxy
		if v, ok := readArgs(a, "proxy="); ok {
			isSocks := false
			if strings.HasPrefix(strings.ToLower(v), "socks5") {
				isSocks = true
			}
			v = strings.Replace(v, "socks5://", "", -1)
			err := h.SetProxy(v, isSocks)
			if err != nil {
				return err
			}
		}
		//secret
		if v, ok := readArgs(a, "secret="); ok {
			h.SetSecret(v)
		}

		//api key
		if v, ok := readArgs(a, "apikey="); ok {
			h.SetAPIKey(v)
		}

		if h.GetAPIKey() == "" || h.GetSecret() == "" {
			return errors.New("please set api key or secret key")
		}

		if strings.ToLower(a[1]) == "btcusdt" {
			e := st.Make(a[1])
			if e != nil {
				return e
			}
		}
		return nil
	}

	if len(a) > 0 && strings.HasPrefix(strings.ToLower(a[0]), "tehran") {

		if v, ok := readArgs(a, "cachepath="); ok {
			h.SetRootCache(v)
		}

		if _, ok := readArgs(a, "-synclist"); ok {
			var dbLock sync.Mutex
			e := av.SyncStockList(&dbLock)
			if e != nil {
				return errors.New(fmt.Sprintf("tehran failed: %v", e))
			}
			return nil
		}

		isreadFromLast := false
		if _, ok := readArgs(a, "-readfromlast"); ok {
			isreadFromLast = true
		}

		if _, ok := readArgs(a, "-stockList"); ok {
			var dbLock sync.Mutex
			e := av.OutStockList(&dbLock)
			if e != nil {
				return errors.New(fmt.Sprintf("tehran failed: %v", e))
			}
			return nil
		}

		e := avardMainProcess(isreadFromLast)
		if e != nil {
			return errors.New(fmt.Sprintf("tehran failed: %v", e))
		}
		return nil
	}

	s := fmt.Sprintf("Help args : [crypto] [BTCUSDT]\nHelp args : [tehran] [src dir path ] [dst dir path]")
	return errors.New(s)

}

func main() {

	e := commands(os.Args[1:])
	if e != nil {
		fmt.Printf(e.Error())
		return
	}
	fmt.Println("finished :)")
}
