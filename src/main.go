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

func init() {
	if h.GetVerbose() {
		fmt.Printf("app init \n")
	}
	//err := h.SetProxy("127.0.0.1:9150", true)
	//err := h.SetProxy("https://127.0.0.1:5051", false) // psiphon
	/*	if err != nil {
		return
	}*/
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
func help() {
	var s string

	s = fmt.Sprintf("\n:::::::::::::::: HELP :::::::::::::::")
	s += fmt.Sprintf("\n:::Make license\t\t-> license make xxxx@gmail.com \n")
	s += fmt.Sprintf("\n:::Mak acti\t\t-> license activated [path license] [days number] [number items stock ] [-type=cft c=crypto , f=forex , t=tehran]  \n")
	s += fmt.Sprintf(":::Crypto Store\t\t-> crypto [params]\n " +
		"\t\t\t\t\tset proxy\t\t\t\t\t\t[proxy=http://xxx.xxx.xxx.xxx:port]\n " +
		"\t\t\t\t\tset cache path\t\t\t\t\t[cachepath=]\n " +
		"\t\t\t\t\tinterval time\t\t\t\t\t[timer=(value) as minuth]\n " +
		"\t\t\t\t\tAdd Stock in Watch List\t\t\t[- add (stock Name) (stock Code)  ]\n " +
		"\t\t\t\t\tget Last data\t\t\t\t\t[-l]\n")

	s += fmt.Sprintf("\n:::Tehran Store\t\t-> tehran [params]\n " +
		"\t\t\t\t\tset proxy\t\t\t\t\t\t[proxy=http://xxx.xxx.xxx.xxx:port]\n " +
		"\t\t\t\t\tset cache path\t\t\t\t\t[cachepath=]\n " +
		"\t\t\t\t\tinterval time\t\t\t\t\t[timer=(value) as minuth]\n " +
		"\t\t\t\t\tAdd Stock in Watch List\t\t\t[ -add (stock Name) (stock Code)  ]\n " +
		"\t\t\t\t\ttemp Watch List\t\t\t\t\t[ -tempwatchlist  ]\n " +
		"\t\t\t\t\tOut All stock list in txt file \t[ -o  ]\n " +
		"\t\t\t\t\tget Last data\t\t\t\t\t[ -l ]\n")

	fmt.Println(s)
}

//_______________________________________________________________________________ COMMANDS
func commands(a []string) error {
	var g *int64 = nil
	li := h.LicenseGen{}
	if _, ok := readArgs(a, "-v"); ok {
		h.SetVerbose(true)
	}
	if len(a) > 0 && strings.ToLower(a[0]) == "-h" {
		help()
		return nil
	}
	//_________________________________________________________________________________________________ LICENSE
	if len(a) > 0 && strings.ToLower(a[0]) == "license" {

		if strings.ToLower(a[1]) == "make" {
			e := li.MakeLicense(a[2])
			if e != nil {
				return e
			}
			return nil
		}
		//(license_path string, days int, items_num int, is_cryto bool, is_tehran bool, is_forex bool)

		if strings.ToLower(a[1]) == "activated" {
			var tehran, forex, crypto bool
			if v, ok := readArgs(a, "type="); ok {
				if len(v) == 0 {
					return errors.New("add type [type=cft] c= cryoti f= forex t=tehran")
				}
				for _, char := range v {
					if char == 'c' {
						crypto = true
					} else if char == 'f' {
						forex = true
					} else if char == 't' {
						tehran = true
					}
				}

			} else
			{
				return errors.New("add type [type=cft] c= cryoti f= forex t=tehran")
			}
			e := li.MakeActivate(a[2], h.ToINT32(a[3]), h.ToINT32(a[4]), crypto, tehran, forex)
			if e != nil {
				return e
			}
			return nil
		}

		return errors.New("Syntax Error")
	}

	result := li.Validation()
	if result != nil {
		if h.GetVerbose() {
			fmt.Println(result)
		}
		fmt.Println("License Not Valid ")
		return nil
	}

	//_________________________________________________________________________________________________ CRYPTO
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
			return errors.New(fmt.Sprintf("config read failed [%v] ", er))
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


	//_________________________________________________________________________________________________ BINANCE
	if len(a) > 0 && strings.HasPrefix(strings.ToLower(a[0]), "binance") {
		biance := av.NewBinance(h.OneFolder)
		if v, ok := readArgs(a, "cachepath="); ok {
			h.SetRootCache(v)
		}

		if v, ok := readArgs(a, "timer="); ok {

			i, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				g = &i
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
			return errors.New(fmt.Sprintf("config read failed [%v] ", er))
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

	//t:=	h.UnixTimeToTime(1580720399000);
	//fmt.Printf("%v",t);
	e := commands(os.Args[1:])
	if e != nil {
		fmt.Println(":::::::::::::::: ERROR :::::::::::::::")
		fmt.Printf(e.Error())
		return
	}
	if h.GetVerbose() {
		fmt.Println("finished :)")
	}
}
