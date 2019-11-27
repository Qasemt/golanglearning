package avardstock

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	. "github.com/qasemt/helper"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

//var lockeList []dbItem

func downloadAsset(assetCode string, isIndex bool, item TimeRange, timefram ETimeFrame, tc ETypeChart) ([]StockFromWebService, error) {
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
	var isAssetStr string = "asset" //asset / index
	if tc == Adj {
		typechart = "%3Atype1"
	}
	if isIndex == true {
		isAssetStr = "index"
	}
	//var raws []interface{}
	var raws []stocktemp
	var itemsFinal []StockFromWebService
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange."+isAssetStr+"%3A"+assetCode+"%3Areal_close"+typechart+"&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &raws)

	if err != nil {
		return nil, err
	}

	if _rawKlines == nil {
		return nil, errors.New("downloadAsset failed ...")
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
func GetDateRangeYears(duration time.Duration, end time.Time) []TimeRange {
	day_rang := []TimeRange{}
	start := end.Add(duration)
	diff := end.Sub(start).Hours() / 8760 //8760 hour = years
	diff = diff + 1
	for i := 0; i <= int(diff); i++ {
		var tt = start.AddDate(i, 0, 0)
		var d1 TimeRange
		d1.File_name = TimeToString(tt, "yyyymmdd") + ".csv"
		y, _, _ := tt.Date()
		d1.Begin = time.Date(y, 1, 1, 0, 0, 0, 0, tt.Location())
		d1.End = time.Date(y, 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())
		day_rang = append(day_rang, d1)
	}
	return day_rang
}
func SyncDb(assetCode string, frame ETimeFrame) error {
	var db *gorm.DB

	db, _, er := DatabaseInit(assetCode, frame.ToString(), db)
	if er != nil {
		return er
	}
	db.Close()
	return nil
}
func Make(wg *sync.WaitGroup, dbLock *sync.Mutex, readfromLast bool, assetCode string, assetNameEn string, isIndex bool, duration time.Duration, end time.Time, timeFrame ETimeFrame, tc ETypeChart) error {

	defer wg.Done()

	var db *gorm.DB
	var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::
	var last = StockFromWebService{
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

	var dbname string = assetCode
	if isIndex {
		dbname = fmt.Sprintf("%si", assetCode)
	}
	db, fullPath, er := DatabaseInit(dbname, timeFrame.ToString(), db)
	if er != nil {
		return errors.New(fmt.Sprintf("err:%v %v", er, fullPath))
	}
	//_________________

	/*	var isFind bool = false
		var lock1 *sync.Mutex = dbLock
		for _, g := range lockeList {
			if g.p == fullPath {
				isFind = true
				lock1 = g.mutex
			}

		}
		if isFind == false {
			lockeList = append(lockeList, dbItem{db: db, p: fullPath, mutex: dbLock})
		}*/

	defer closeMyDb(db)

	//::::::::::::::::::::::::::::::::::::::::: Get LOOP FROM WEB SERVICE
	var times []TimeRange
	var it = TimeRange{}
	//var itemsFinal []StockItem
	if readfromLast {
		//::::::::::::::::::::::::::::::::::::::::: Get LAst RECORD FROM DATABASE
		e := getLastRecord(db, dbLock, assetCode, timeFrame.ToMinuth(), tc, &last)
		if e != nil {
			return e
		}
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
		times = append(times, it)
	} else
	{
		t := GetDateRangeYears(duration, end)
		times = append(times, t...)
	}

	var itemsRaws []StockFromWebService
	for _, h := range times {
		raws, e := downloadAsset(assetCode, isIndex, h, timeFrame, tc)
		if e != nil {
			return e
		}
		itemsRaws = append(itemsRaws, raws...)
	}

	//::::::::::::::::::::::::::::::::::::::::: INSERT TO DATABASE
	{
		fmt.Println("Type", tc.ToTypeChartStr(), "asset ", assetNameEn, "time frame ", timeFrame.ToString(), "load from net : ", len(itemsRaws))
		if len(itemsRaws) > 0 {
			InsertStocks(db, dbLock, isIndex, itemsRaws, assetCode, timeFrame, tc)
			//if err != nil {
			//	return errors.New(fmt.Sprintf("Insert Stocks is fialed: %v ",err))
			//}
		}
	}
	//::::::::::::::::::::::::::::::::::::::::: LOAD FROM DATABASE AND OUT TO CSV
	{
		itemsRaw, err := getRecordesStock(db, dbLock, assetCode, timeFrame, tc)
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
			var dirCachePath string
			if tc == Normal {
				dirCachePath = path.Join(GetRootCache(), "tehran", "normal", timeFrame.ToString())
			} else {
				dirCachePath = path.Join(GetRootCache(), "tehran", "Adjusted", timeFrame.ToString())
			}
			var fileName string = ""
			if tc == Normal {
				fileName = fmt.Sprintf("%v_%v.csv", assetNameEn, strings.ToLower(timeFrame.ToString2()))
			} else {
				fileName = fmt.Sprintf("%v_%v_%v.csv", assetNameEn, strings.ToLower(timeFrame.ToString2()), "a")
			}

			if !OutToCSVFile(itemsFinal, dirCachePath, fileName, true) {
				return errors.New("get asset daily >>> out to csv failed")
			}
		}
		//fmt.Println("final :", len(itemsFinal))
	}
	return nil
}
func ReadJsonWatchList() (*WatchListItem, error) {
	var list WatchListItem
	watchPath := path.Join(GetRootCache(), "watchList.json")
	if !IsExist(watchPath) {
		//return nil, errors.New(fmt.Sprintf("watch list not found : %v", watchPath))
		fmt.Printf(fmt.Sprintf("watch list not found : %v ,create default Watch list ", watchPath))
		CreateWatchList(GetRootCache())
	}

	jsonFile, err := os.Open(watchPath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	e := json.Unmarshal(byteValue, &list)
	if e != nil {
		return nil, e
	}
	return &list, nil
}
func SyncStockList(dbLock *sync.Mutex) error {

	var db1 *gorm.DB
	//var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::;
	db, _, er := DatabaseInit("main", "", db1)
	if er != nil {
		return er
	}
	type NemadAvardRaw struct {
		TypeId      string `json:"type_id"`
		Type        string `json:"type"`
		EntityId    string `json:"entity_id"`
		EntityType  string `json:"entity_type"`
		TradeSymbol string `json:"trade_symbol"`
		Title       string `json:"title"`
	}
	type assetList struct {
		Data  []NemadAvardRaw ` json:"data"`
		Error string          ` json:"error"`
		Meta  string          ` json:"meta"`
	}

	var rawsAsset assetList
	var rawsIndex assetList

	errAsset := GetJson("https://rahavard365.com/api/search/items?type=asset", &rawsAsset)

	if errAsset != nil {
		return errAsset
	}
	errIndex := GetJson("https://rahavard365.com/api/search/items?type=index", &rawsIndex)

	if errIndex != nil {
		return errIndex
	}

	if rawsAsset.Data == nil || rawsIndex.Data == nil {
		return errors.New("SyncStockList failed ... ")
	}

	if len(rawsAsset.Data) == 0 || len(rawsIndex.Data) == 0 {
		return errors.New("SyncStockList -> data from net is empty  ... ")
	}
	var Items []NemadAvard
	for i := 0; i < len(rawsAsset.Data); i++ {
		it := rawsAsset.Data[i]
		n := NemadAvard{}
		n.Title = it.Title
		n.TradeSymbol = it.TradeSymbol
		n.EntityType = it.EntityType
		n.EntityId = ToINT64(it.EntityId)
		n.Type = it.Type
		n.TypeId = it.TypeId
		Items = append(Items, n)
	}

	for i := 0; i < len(rawsIndex.Data); i++ {
		it := rawsIndex.Data[i]
		n := NemadAvard{}
		n.Title = it.Title
		n.TradeSymbol = it.TradeSymbol
		n.EntityType = it.EntityType
		n.EntityId = ToINT64(it.EntityId)
		n.Type = it.Type
		n.TypeId = it.TypeId
		Items = append(Items, n)
	}

	e1 := InsertAssetInfoFromAvard(db, dbLock, Items)
	if e1 != nil {
		return e1
	}

	return nil
}

func OutStockList(dbLock *sync.Mutex) error {
	return nil
}
