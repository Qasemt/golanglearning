package avardstock

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	. "github.com/qasemt/helper"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

var cachePath = "d:/cache2/tehran/"

type stocktemp struct {
	Time float64 ` json:"time"`
	O    float64 `json:"open"`
	H    float64 `json:"high"`
	L    float64 `json:"low"`
	C    float64 `json:"close"`
	V    float64 `json:"volume"`
}
type dbItem struct {
	db    *gorm.DB
	p     string
	mutex *sync.Mutex
}

func downloadAsset(assetName string, item TimeRange, timefram ETimeFrame, tc ETypeChart) ([]StockFromWebService, error) {
	var _rawKlines = []StockFromWebService{}
	startStr := strconv.FormatInt(item.Begin.Unix(), 10)
	endStr := strconv.FormatInt(item.End.Unix(), 10)
	var frame string
	if timefram == D1 {
		frame = "D"
	} else if timefram == M15 {
		frame = "15"
	} else if timefram == H1 {
		frame = "60"
	} else if timefram == H2 {
		frame = "120"
	} else if timefram == H4 {
		frame = "240"
	}
	var typechart string = ""
	if tc == Adj {
		typechart = "%3Atype1"
	}
	//var raws []interface{}
	var raws []stocktemp
	var itemsFinal []StockFromWebService
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange.asset%3A66%3Areal_close"+typechart+"&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &raws)

	if err != nil {
		return nil, err
	}

	if _rawKlines == nil {
		return nil, errors.New("downloadAsset failed ... binance block link")
	}

	for _, k := range raws {
		var v StockFromWebService
		v.Time = int64(k.Time)
		v.O = k.O
		v.H = k.H
		v.L = k.L
		v.C = k.C
		v.V = k.V
		itemsFinal = append(itemsFinal, v)

	}

	return itemsFinal, nil
}
func closeMyDb(d *gorm.DB) {
	if d != nil {
		(*d).Close()

	}
}

var lockelist []dbItem

func Make(wg *sync.WaitGroup, lock *sync.Mutex,assetCode string, assetName string, duration time.Duration, end time.Time, timeFrame ETimeFrame, tc ETypeChart) error {
	defer wg.Done()
	var db *gorm.DB
	var dirCachePath string = "./d/"
	var fullPath = path.Join(dirCachePath, fmt.Sprintf("%v_%v.bin", assetCode, timeFrame.ToString()))
	if _, err := os.Stat(dirCachePath); os.IsNotExist(err) {
		os.MkdirAll(dirCachePath, os.ModePerm)
	}
	//:::::::::::::::::::::::::::::::::::::::::;

	db, er := DatabaseInit(fullPath, nil)
	if er != nil {
		return er
	}
	//_________________
	var isFind bool =false
	var lock1 *sync.Mutex =lock
	for _, g := range lockelist {
	if g.p == fullPath{
		isFind =true
		lock1=g.mutex
	}

	}
	if isFind==false {
		lockelist = append(lockelist, dbItem{db: db, p: fullPath, mutex: lock})
	}


	//:::::::::::::::::::::::::::::::::::::::::;
	defer closeMyDb(db)
	var last StockFromWebService
	last = StockFromWebService{
		ID:        0,
		AssetId:   "",
		TimeFrame: 0,
		TypeChart: 0,
		Time:      0,
		O:         0,
		H:         0,
		L:         0,
		C:         0,
		V:         0,
	}
	e := getLastRecord(db, assetCode, timeFrame.ToMinuth(), tc, &last)
	if e != nil {
		return e
	}
	//fmt.Println(last.TOString())

	var it = TimeRange{}
	//var itemsFinal []StockItem
	if last.ID == 0 {
		it.Begin = end.Add(duration)
	} else {
		//it.Begin = time.Unix(0, int64(last.Time)*int64(time.Millisecond))
		if last.Time == 0 {
			return errors.New("last time not valid ")
		}

		it.Begin = time.Unix(0, last.Time*int64(time.Millisecond))
	}
	it.End = end
	fmt.Println("->", it)
	itemsRaws, e := downloadAsset(assetCode, it, timeFrame, tc)

	if e != nil {
		return e
	}

	//:::::::::::::::::::::::::::
	println("load net : ", len(itemsRaws))
	if len(itemsRaws) > 0 {
		InsertStocks(db,lock1, itemsRaws, assetCode, timeFrame, tc)
		//if err != nil {
		//	return errors.New(fmt.Sprintf("Insert Stocks is fialed: %v ",err))
		//}
	}
	itemsRaw, err := getRecordes(db, assetCode, timeFrame, tc)
	if err != nil {
		return errors.New(fmt.Sprintf("get Stocks is failed: %v ", err))
	}
	var itemsFinal []StockItem
	for _, k := range itemsRaw {
		var v StockItem
		time1 := time.Unix(0, int64(k.Time)*int64(time.Millisecond))
		v.D = UnixTimeStrToFormatDT(time1, true)
		v.T = UnixTimeStrToFormatDT(time1, false)

		v.O = k.O
		v.H = k.H
		v.L = k.L
		v.C = k.C
		v.V = k.V
		itemsFinal = append(itemsFinal, v)
	}

	if len(itemsFinal) > 0 {
		dir_cache_path := path.Join(cachePath, assetCode)
		var fileName string = ""
		if tc == Normal {
			fileName = fmt.Sprintf("%v_%v.csv", assetName, timeFrame.ToString())
		} else {
			fileName = fmt.Sprintf("%v_%v_adj.csv", assetName, timeFrame.ToString())
		}

		if !OutToCSVFile(itemsFinal, dir_cache_path, fileName, true) {
			return errors.New("get asset daily >>> out to csv failed")
		}
	}
	fmt.Println("final :", len(itemsFinal))
	return nil
}
