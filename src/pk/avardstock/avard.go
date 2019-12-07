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

func floatFromString(raw interface{}) (float64, error) {
	str, ok := raw.(string)
	if !ok {
		//	return 0, error(fmt.Sprintf("unable to parse, value not string: %T", raw))
		return 0, nil
	}
	flt, err := strconv.ParseFloat(str, 64)
	if err != nil {
		//	return 0, errors.Wrap(err, fmt.Sprintf("unable to parse as float: %s", str))
		return 0, nil

	}
	return flt, nil
}
func timeFromUnixTimestampFloat(raw interface{}) (time.Time, error) {
	ts, ok := raw.(float64)
	if !ok {
		return time.Time{}, nil
	}
	return time.Unix(0, int64(ts)*int64(time.Millisecond)), nil
}

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
type IStockProvider interface {
	make(sq StockQuery) error
	downloadAsset(sq StockQuery, item TimeRange) ([]StockFromWebService, error)
	/*
		closeMyDb(d *gorm.DB)
		getDateRangeYears(duration time.Duration, end time.Time) []TimeRange
		SyncDb(wl *WatchListItem) error
		ReadJsonWatchList() (*WatchListItem, error)
		SyncStockList(dbLock *sync.Mutex) error
		OutStockList(dbLock *sync.Mutex)
		avardAssetProcess(parentWaitGroup *sync.WaitGroup, readFromLast bool, assetCode string, nameEn string, isIndex bool, provider EProvider) error

	*/
	Run(readfromLast bool) error
}
type StockProvider struct {
	IStockProvider
	Provider EProvider
}
type TehranLoader struct {
	StockProvider
}
type BinanceLoader struct {
	StockProvider
}

//::::::::::::::::::::::::::::::::::::
func NewTehran() *TehranLoader {
	t := TehranLoader{StockProvider{}}

	t.StockProvider.IStockProvider = &t
	t.Provider = Avard
	return &t
}
func NewBinance() *BinanceLoader {
	t := BinanceLoader{StockProvider{}}

	t.StockProvider.IStockProvider = &t
	t.Provider = Binance
	return &t
}

func (a TehranLoader) downloadAsset(sq StockQuery, item TimeRange) ([]StockFromWebService, error) {
	var _rawKlines = []StockFromWebService{}
	startStr := strconv.FormatInt(item.Begin.Unix(), 10)
	endStr := strconv.FormatInt(item.End.Unix(), 10)
	var frame string
	if sq.TimeFrame == D1 {
		frame = "D"
	} else if sq.TimeFrame == M15 {
		frame = "15"
	} else if sq.TimeFrame == H1 {
		frame = "60"
	} else if sq.TimeFrame == H2 {
		frame = "120"
	} else if sq.TimeFrame == H4 {
		frame = "240"
	}
	var typechart string = ""
	var isAssetStr string = "asset" //asset / index
	if sq.TypeChart == Adj {
		typechart = "%3Atype1"
	}
	if sq.IsIndex == true {
		isAssetStr = "index"
	}
	//var raws []interface{}
	var raws []stocktemp
	var itemsFinal []StockFromWebService
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange."+isAssetStr+"%3A"+sq.AssetCode+"%3Areal_close"+typechart+"&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &raws)

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
func (a BinanceLoader) downloadAsset(sq StockQuery, item TimeRange) ([]StockFromWebService, error) {
	var _rawKlines = []StockFromWebService{}
	startStr := strconv.FormatInt(UnixMilli(item.Begin), 10)
	endStr := strconv.FormatInt(UnixMilli(item.End), 10)
	rawKlines := [][]interface{}{}
	var itemsFinal []StockFromWebService
	err := GetJsonBin("api/v3/klines?symbol="+sq.AssetCode+"&interval="+sq.TimeFrame.ToString()+"&startTime="+startStr+"&endTime="+endStr, &rawKlines)

	if err != nil {
		return nil, err
	}

	if _rawKlines == nil {
		return nil, errors.New("downloadAsset failed ...")
	}

	for _, k := range rawKlines {
		var v StockFromWebService
		ts, _ := k[0].(float64)
		v.Time = int64(ts)
		open, _ := floatFromString(k[1])
		v.O = open

		high, _ := floatFromString(k[2])
		v.H = high

		low1, _ := floatFromString(k[3])
		v.L = low1

		close, _ := floatFromString(k[4])
		v.C = close

		volume, _ := floatFromString(k[5])
		v.V = volume

		itemsFinal = append(itemsFinal, v)

	}

	return itemsFinal, nil
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func (a StockProvider) make(sq StockQuery) error {

	defer sq.WaitGroupobj.Done()

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

	var dbname string = sq.AssetCode
	if sq.IsIndex {
		dbname = fmt.Sprintf("%si", sq.AssetCode)
	}
	db, fullPath, er := DatabaseInit(dbname, sq.TimeFrame.ToString(), db)
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

	defer a.closeMyDb(db)

	//::::::::::::::::::::::::::::::::::::::::: Get LOOP FROM WEB SERVICE
	var times []TimeRange
	var it = TimeRange{}
	//var itemsFinal []StockItem
	if sq.ReadfromLast {
		//::::::::::::::::::::::::::::::::::::::::: Get LAst RECORD FROM DATABASE
		e := getLastRecord(db, sq.DBLock, sq.AssetCode, sq.TimeFrame.ToMinuth(), sq.TypeChart, &last)
		if e != nil {
			return e
		}
		if last.ID == 0 {
			it.Begin = sq.EndTime.Add(sq.Duration)
		} else {
			//it.Begin = time.Unix(0, int64(last.Time)*int64(time.Millisecond))
			if last.Time == 0 {
				return errors.New("last time not valid ")
			}
			it.Begin = time.Unix(0, last.Time*int64(time.Millisecond))
		}
		it.End = sq.EndTime
		times = append(times, it)
	} else
	{
		t := a.getDateRangeYears(sq.Duration, sq.EndTime)
		times = append(times, t...)
	}

	var itemsRaws []StockFromWebService
	if a.Provider == Avard {
		for _, h := range times {
			raws, e := a.downloadAsset(sq, h)
			if e != nil {
				return e
			}
			itemsRaws = append(itemsRaws, raws...)
		}

	} else if a.Provider == Binance {
		for _, h := range times {
			l := a.getDateRangeBy500Hours(h.Begin, h.End, sq.TimeFrame)

			for _, h1 := range l {
				raws, e := a.downloadAsset(sq, h1)
				if e != nil {
					return e
				}
				itemsRaws = append(itemsRaws, raws...)
			}

		}
	} else {
		return errors.New("no selected")
	}

	//::::::::::::::::::::::::::::::::::::::::: INSERT TO DATABASE
	{
		fmt.Println(a.Provider.ToString(), "->", "Type", sq.TypeChart.ToTypeChartStr(), "asset ", sq.AssetNameEn, "time frame ", sq.TimeFrame.ToString(), "load from net : ", len(itemsRaws))
		if len(itemsRaws) > 0 {
			InsertStocks(db, sq.DBLock, sq.IsIndex, itemsRaws, sq.AssetCode, sq.TimeFrame, sq.TypeChart)
			//if err != nil {
			//	return errors.New(fmt.Sprintf("Insert Stocks is fialed: %v ",err))
			//}
		}
	}
	//::::::::::::::::::::::::::::::::::::::::: LOAD FROM DATABASE AND OUT TO CSV
	{
		itemsRaw, err := getRecordesStock(db, sq.DBLock, sq.AssetCode, sq.TimeFrame, sq.TypeChart)
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
			var fileName string = ""
			switch a.Provider {
			case Avard:
				{
					if sq.TypeChart == Normal {
						dirCachePath = path.Join(GetRootCache(), "tehran", "normal", sq.TimeFrame.ToString())
					} else {
						dirCachePath = path.Join(GetRootCache(), "tehran", "Adjusted", sq.TimeFrame.ToString())
					}
					if sq.TypeChart == Normal {
						fileName = fmt.Sprintf("%v_%v.csv", sq.AssetNameEn, strings.ToLower(sq.TimeFrame.ToString2()))
					} else {
						fileName = fmt.Sprintf("%v_%v_%v.csv", sq.AssetNameEn, strings.ToLower(sq.TimeFrame.ToString2()), "a")
					}
				}
			case Binance:
				{
					dirCachePath = path.Join(GetRootCache(), "crypto", sq.AssetCode)
					fileName = fmt.Sprintf("%v_%v.csv", strings.ToLower(sq.AssetCode), strings.ToLower(sq.TimeFrame.ToString2()))
				}
			}

			if !OutToCSVFile(itemsFinal, dirCachePath, fileName, true) {
				return errors.New("get asset daily >>> out to csv failed")
			}
		}
		//fmt.Println("final :", len(itemsFinal))
	}
	return nil
}

func (a StockProvider) closeMyDb(d *gorm.DB) {
	if d != nil {
		(*d).Close()

	}
}
func (a StockProvider) getDateRangeYears(duration time.Duration, end time.Time) []TimeRange {
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
func (a StockProvider) getDateRangeBy500Hours(start time.Time, end time.Time, frame ETimeFrame) []TimeRange {
	day_rang := []TimeRange{}
	var diff float64
	switch frame {
	case M15:
		diff = (end.Sub(start).Minutes() / 15) / 499
	case H1:
		{
			diff = end.Sub(start).Hours() / 499
		}
	case H2:
		{
			diff = (end.Sub(start).Hours() / 2) / 499 //8760 hour = years
		}
	case H4:
		{
			diff = (end.Sub(start).Hours() / 4) / 499 //8760 hour = years
		}
	case D1:
		{
			var d1 TimeRange
			d1.File_name = TimeToString(start, "yyyymmdd") + ".csv"
			d1.Begin = start
			d1.End = end
			day_rang = append(day_rang, d1)
			return day_rang
		}
	}

	var t1 time.Time
	var t2 time.Time

	for i := 0; i <= int(diff); i++ {
		if i == 0 {
			t1 = start
		} else {
			t1 = t2
		}

		switch frame {
		case M15:
			t2 = t1.Add((time.Minute * 15) * time.Duration(500))
		case H1:
			{
				t2 = t1.Add((time.Hour) * time.Duration(500))
			}
		case H2:
			{
				t2 = t1.Add((time.Hour * 2) * time.Duration(500))
			}
		case H4:
			{
				t2 = t1.Add((time.Hour * 4) * time.Duration(500))
			}
		case D1:
			{

			}
		}
		//::::::::::::::::::::::::::::::::::::::
		//fmt.Println(t1,t2)
		if t1.After(time.Now()) {
			break
		}
		var d1 TimeRange
		d1.File_name = TimeToString(t1, "yyyymmdd") + ".csv"
		d1.Begin = t1
		d1.End = t2
		day_rang = append(day_rang, d1)
	}
	return day_rang
}

/*func (a StockProvider) make(sq StockQuery) error {
	return a.make(sq);
}*/
//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

func (a StockProvider) SyncDb(wl *WatchListItem) error {

	for _, f := range wl.Tehran {
		var dbnametmp string = f.AssetCode
		if f.IsIndex {
			dbnametmp = fmt.Sprintf("%vi", f.AssetCode)
		}

		e := Migrate(dbnametmp, &a)
		if e != nil {
			return e
		}

	}
	//____________
	for _, f := range wl.Crypto {
		var dbnametmp string = f.AssetCode
		if f.IsIndex {
			dbnametmp = fmt.Sprintf("%ti", f.AssetCode)
		}

		e := Migrate(dbnametmp, &a)
		if e != nil {
			return e
		}

	}
	e := Migrate("main", &a)
	if e != nil {
		return e
	}
	fmt.Println("sync ..... done ")
	return nil
}
func (a StockProvider) ReadJsonWatchList() (*WatchListItem, error) {
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
func (a StockProvider) SyncStockList(dbLock *sync.Mutex) error {

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
func (a StockProvider) OutStockList(dbLock *sync.Mutex) error {

	var db1 *gorm.DB
	//var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::;
	db, _, er := DatabaseInit("main", "", db1)
	if er != nil {
		return er
	}
	items, err := GetNemadList(db, dbLock)
	if err != nil {
		return err
	}
	for _, k := range items {
		fmt.Printf("%v %v %v\n",k.EntityType,k.EntityId,k.TradeSymbol)
	}
	return nil
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func (a StockProvider) avardAssetProcess(parentWaitGroup *sync.WaitGroup, readFromLast bool, assetCode string, nameEn string, isIndex bool) error {

	if nameEn == "" || assetCode == "" {
		parentWaitGroup.Done()
		return errors.New("field is empty ")

	}
	var databaseLock sync.Mutex
	var wg sync.WaitGroup
	if a.Provider == Avard {

		if isIndex == true {
			wg.Add(2)
		} else {
			wg.Add(8)
		}

		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H1, TypeChart: Normal})

		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 400), EndTime: time.Now(), TimeFrame: D1, TypeChart: Normal})

		if isIndex == false {
			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H2, TypeChart: Normal})
			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 360), EndTime: time.Now(), TimeFrame: H4, TypeChart: Normal})

			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H1, TypeChart: Adj})
			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H2, TypeChart: Adj})
			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 360), EndTime: time.Now(), TimeFrame: H4, TypeChart: Adj})
			go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: isIndex, Duration: -time.Duration(time.Hour * 24 * 400), EndTime: time.Now(), TimeFrame: D1, TypeChart: Adj})
		}
	} else if a.Provider == Binance {

		wg.Add(5)
		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: false, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: M15, TypeChart: Normal})
		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: false, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H1, TypeChart: Normal})
		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: false, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H2, TypeChart: Normal})
		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: false, Duration: -time.Duration(time.Hour * 24 * 360), EndTime: time.Now(), TimeFrame: H4, TypeChart: Normal})
		go a.make(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, AssetCode: assetCode, AssetNameEn: nameEn, IsIndex: false, Duration: -time.Duration(time.Hour * 24 * 400), EndTime: time.Now(), TimeFrame: D1, TypeChart: Normal})
	} else {
		return errors.New("not selected :( ")
	}
	wg.Wait()
	parentWaitGroup.Done()

	return nil
}
func (a StockProvider) Run(readfromLast bool) error {

	list, e := a.ReadJsonWatchList()

	if e != nil {
		return errors.New(fmt.Sprintf("config not found "))
	}
	var wg sync.WaitGroup
	if a.Provider == Avard {
		wg.Add(len(list.Tehran))
		for _, g := range list.Tehran {
			go a.avardAssetProcess(&wg, readfromLast, g.AssetCode, g.NameEn, g.IsIndex)
			/*	if e != nil {
				return e
			}*/
		}
	} else if a.Provider == Binance {
		wg.Add(len(list.Crypto))
		for _, g := range list.Crypto {
			go a.avardAssetProcess(&wg, readfromLast, g.AssetCode, g.NameEn, g.IsIndex)
			/*	if e != nil {
				return e
			}*/
		}
	}
	wg.Wait()
	return nil
}
