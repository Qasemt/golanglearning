package main

import (
	"errors"
	"fmt"
	av "github.com/qasemt/avardstock"
	b "github.com/qasemt/binance"
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
	err := h.SetProxy("socks://127.0.0.1:9150", true)
	//err := h.SetProxy("http://38.113.170.11:45864", false) // psiphon
	if err != nil {
		return
	}
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
func binanceV2() {
	f := b.MakeCacheHourly("BTCUSDT", h.M15, h.H1, time.Duration(time.Hour*24*1), time.Now())
	fmt.Println(f)

	s := b.MakeCacheBase15M("BTCUSDT", time.Duration(time.Hour*24*10), time.Now())

	fmt.Println(s)

}
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
	wg.Add(len(list))

	for _, g := range list {
		go avardAssetProcess(&wg, readfromLast, g.AssetCode, g.NameEn, g.IsIndex)
		/*	if e != nil {
			return e
		}*/
	}
	wg.Wait()
	return nil
}

//OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO
func commands(a []string) error {
	if len(a) > 0 && strings.ToLower(a[0]) == "crypto" {
		//proxy=socks5://127.0.0.1:9150
		for i := 1; i < len(a); i++ {
			if strings.HasPrefix(strings.ToLower(a[i]), "proxy=") {
				p := strings.Split(a[i], "=")[1]
				p = strings.Trim(p, `"`)
				err := h.SetProxy(p, true)
				if err != nil {
					return err
				}
				break
			}
		}

		if strings.ToLower(a[1]) == "btcusdt" {
			e := st.Make(a[1])
			if e != nil {
				return e
			}
		}

	} else if len(a) > 0 && strings.HasPrefix(strings.ToLower(a[0]), "tehran") {

		for i := 1; i < len(a); i++ {
			if strings.HasPrefix(strings.ToLower(a[i]), "cachepath=") {
				p := strings.Split(a[1], "=")[1]
				p = strings.Trim(p, `"`)
				h.SetRootCache(p)
				break
			}
		}
		for i := 1; i < len(a); i++ {
			if strings.HasPrefix(strings.ToLower(a[i]), "-synclist") {
				var dbLock sync.Mutex
				e := av.SyncStockList(&dbLock)
				if e != nil {
					return errors.New(fmt.Sprintf("tehran failed: %v", e))
				}
				return nil
			}
		}
		isreadFromLast := false
		for i := 1; i < len(a); i++ {
			if strings.HasPrefix(strings.ToLower(a[i]), "-readfromlast") {
				isreadFromLast = true
			}
		}
		for i := 1; i < len(a); i++ {
			if strings.HasPrefix(strings.ToLower(a[i]), "-stockList") {
				var dbLock sync.Mutex
				e := av.OutStockList(&dbLock)
				if e != nil {
					return errors.New(fmt.Sprintf("tehran failed: %v", e))
				}
				return nil
			}
		}

		e := avardMainProcess(isreadFromLast)
		if e != nil {
			return errors.New(fmt.Sprintf("tehran failed: %v", e))
		}

	} else {
		s := fmt.Sprintf("Help args : [crypto] [BTCUSDT]\nHelp args : [tehran] [src dir path ] [dst dir path]")
		return errors.New(s)
	}
	return nil
}

func main() {
	//avardSync()
	//
	e := commands(os.Args[1:])

	if e != nil {
		fmt.Printf(e.Error())
		return
	}

	fmt.Println("finished :)")

}
