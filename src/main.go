package main

import (
	"errors"
	"fmt"
	av "github.com/qasemt/avardstock"
	h "github.com/qasemt/helper"
	"os"
	"strconv"
	"strings"
	"sync"
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
	var g *int64 = nil
	if len(a) > 0 && strings.ToLower(a[0]) == "crypto" {
		binance := av.NewBinance(h.OneFolder)
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

		//0000000000000
		list, er := binance.ReadJsonWatchList()
		if er != nil {
			return errors.New(fmt.Sprintf("config not found "))
		}
		k := binance.SyncDb(list)
		if k != nil {
			return errors.New(fmt.Sprintf("sync db failed."))
		}
		h.SetSecret(list.Apikey)
		h.SetAPIKey(list.Secret)

		if h.GetAPIKey() == "" || h.GetSecret() == "" {
			return errors.New("please set api key or secret key")
		}

		return nil
	}

	if len(a) > 0 && strings.HasPrefix(strings.ToLower(a[0]), "tehran") {
		tehran := av.NewTehran(h.OneFolder)
		if v, ok := readArgs(a, "cachepath="); ok {
			h.SetRootCache(v)
		}


		if v, ok := readArgs(a, "timer="); ok {

			i, err :=  strconv.ParseInt(v, 10, 64)
			if err == nil {
				g =&i
			} else {
				return fmt.Errorf("timer not valid %v", err)
			}
		}

		isSeq := false
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

		if _, ok := readArgs(a, "-list"); ok {
			var dbLock sync.Mutex
			e := tehran.SyncStockList(&dbLock)
			if e != nil {
				return errors.New(fmt.Sprintf("tehran failed: %v", e))
			}
			return nil
		}
		if _, ok := readArgs(a, "-seq"); ok {
			isSeq = true
		}

		isreadFromLast := false
		if _, ok := readArgs(a, "-l"); ok {
			isreadFromLast = true
		}

		if _, ok := readArgs(a, "-o"); ok {
			var dbLock sync.Mutex
			e := tehran.OutStockList(&dbLock)
			if e != nil {
				return errors.New(fmt.Sprintf("tehran failed: %v", e))
			}
			return nil
		}
		//0000000000000
		wlist, er := tehran.ReadJsonWatchList()
		if er != nil {
			return errors.New(fmt.Sprintf("config not found "))
		}
		k := tehran.SyncDb(wlist)
		if k != nil {
			return errors.New(fmt.Sprintf("sync db failed."))
		}

		e := tehran.Run(isreadFromLast, isSeq, g)
		if e != nil {
			return errors.New(fmt.Sprintf("tehran failed: %v", e))
		}
		return nil
	}

	if len(a) > 0 && strings.HasPrefix(strings.ToLower(a[0]), "binance") {
		biance := av.NewBinance(h.OneFolder)
		if v, ok := readArgs(a, "cachepath="); ok {
			h.SetRootCache(v)
		}

		if v, ok := readArgs(a, "timer="); ok {

			i, err :=  strconv.ParseInt(v, 10, 64)
			if err == nil {
				g =&i
			} else {
				return fmt.Errorf("timer not valid %v", err)
			}
		}
		isreadFromLast := false
		isSeq := false
		if _, ok := readArgs(a, "-l"); ok {
			isreadFromLast = true
		}
		if _, ok := readArgs(a, "-seq"); ok {
			isSeq = true
		}

		if _, ok := readArgs(a, "-list"); ok {
			var dbLock sync.Mutex
			e := biance.OutStockList(&dbLock)
			if e != nil {
				return errors.New(fmt.Sprintf("tehran failed: %v", e))
			}
			return nil
		}
		//0000000000000
		list, er := biance.ReadJsonWatchList()
		if er != nil {
			return errors.New(fmt.Sprintf("config not found "))
		}
		k := biance.SyncDb(list)
		if k != nil {
			return errors.New(fmt.Sprintf("sync db failed."))
		}

		e := biance.Run(isreadFromLast, isSeq, g)
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
