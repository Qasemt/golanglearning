package avardstock

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	. "github.com/qasemt/helper"
	"golang.org/x/sync/semaphore"
	_ "golang.org/x/sync/semaphore"
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


type Rahavard_Data struct {
	Data []struct {

		Time    float64     `json:"time"`
		O    float64     `json:"open"`
		H    float64     `json:"high"`
		L     float64     `json:"low"`
		C   float64     `json:"close"`
		V  float64         `json:"volume"`
	} `json:"data"`
	NextTime interface{} `json:"nextTime"`
	NoData   interface{} `json:"noData"`
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
	Run(readfromLast bool, isSeq bool, timer_minute *int64) error
}
type StockProvider struct {
	IStockProvider
	Provider        EProvider
	FolderStoreMode EFolderStoreMode
	IsSeqRunProcess bool
	_WatchListItem  *WatchListItem
	HttpLock        sync.Mutex
	sem             *semaphore.Weighted
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func (a StockProvider) procMake(sq StockQuery) {
	if a.IsSeqRunProcess {
		a.make(sq)
	} else {
		sq.WaitGroupobj.Add(1)
		go a.make(sq)
	}
}
func (a StockProvider) make(sq StockQuery) error {

	if !a.IsSeqRunProcess {
		defer sq.WaitGroupobj.Done()
	}
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

	var dbname string = sq.Stock.AssetCode
	if sq.Stock.IsIndex {
		dbname = fmt.Sprintf("%si", sq.Stock.AssetCode)
	}
	db, fullPath, er := DatabaseInit(dbname, sq.TimeFrame.ToString())

	if er != nil {
		return errors.New(fmt.Sprintf("err:%v %v", er, fullPath))
	}
	defer a.closeMyDb(db)
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
		e := getLastRecord(db, sq.DBLock, sq.Stock.AssetCode, sq.TimeFrame.ToMinuth(), sq.TypeChart, &last)
		if GetVerbose() {
			fmt.Printf("last record -> timeframe : %v  time : %v\n", sq.TimeFrame.ToString2(), UnixTimeToTime(last.Time).ToString());
		}
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
			it.Begin = time.Unix(0, (last.Time)*int64(time.Millisecond))
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

			if GetVerbose() {
				fmt.Printf("times range -> %v\n", h.ToString())
			}
			l := a.getDateRangeBy500Hours(h.Begin, h.End, sq.TimeFrame)

			for _, h1 := range l {
				raws, e := a.downloadAsset(sq, h1)
				if e != nil {
					fmt.Printf("make()-> %v | %v | %v\n", sq.Stock.NameEn, sq.TimeFrame.ToString2(), e)
					return e
				}
				itemsRaws = append(itemsRaws, raws...)
			}

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
		fmt.Println(a.Provider.ToString(), "->", "Type", sq.TypeChart.ToTypeChartStr(), "asset ", sq.Stock.NameEn, "time frame ", sq.TimeFrame.ToString(), "load from net : ", len(itemsRaws))
		if len(itemsRaws) > 0 {
			InsertStocks(db, sq.DBLock, last, sq.Stock.IsIndex, itemsRaws, sq.Stock.AssetCode, sq.TimeFrame, sq.TypeChart)
			//if err != nil {
			//	return errors.New(fmt.Sprintf("Insert Stocks is fialed: %v ",err))
			//}
		}
	}
	//::::::::::::::::::::::::::::::::::::::::: LOAD FROM DATABASE AND OUT TO CSV
	{
		itemsRaw, err := getRecordesStock(db, sq.DBLock, sq.Stock.AssetCode, sq.TimeFrame, sq.TypeChart)
		if err != nil {
			return errors.New(fmt.Sprintf("get Stocks is failed: %v ", err))
		}
		var itemsFinal []StockItem
		for _, k := range itemsRaw {
			var v StockItem
			time1 := time.Unix(0, int64(k.Time)*int64(time.Millisecond))
			v.D = UnixTimeStrToFormatDT(time1, true, sq.TimeFrame)
			v.T = UnixTimeStrToFormatDT(time1, false, sq.TimeFrame)

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
					if a.FolderStoreMode == ByTimeFrame {
						if sq.TypeChart == Normal {
							dirCachePath = path.Join(GetRootCache(), "tehran", "normal", sq.TimeFrame.ToString())
						} else {
							dirCachePath = path.Join(GetRootCache(), "tehran", "Adjusted", sq.TimeFrame.ToString())
						}

					} else {
						dirCachePath = path.Join(GetRootCache(), "tehran")
					}
					if sq.TypeChart == Normal {
						fileName = fmt.Sprintf("%v_%v.csv", sq.Stock.NameEn, strings.ToLower(sq.TimeFrame.ToString2()))
					} else {
						fileName = fmt.Sprintf("%v_%v_%v.csv", sq.Stock.NameEn, strings.ToLower(sq.TimeFrame.ToString2()), "a")
					}
				}
			case Binance:
				{
					if a.FolderStoreMode == ByTimeFrame {
						dirCachePath = path.Join(GetRootCache(), "crypto", sq.Stock.AssetCode)

					} else {
						dirCachePath = path.Join(GetRootCache(), "crypto")
					}
					fileName = fmt.Sprintf("%v_%v.csv", strings.ToLower(sq.Stock.AssetCode), strings.ToLower(sq.TimeFrame.ToString2()))
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

//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: dorost kardan db stock ha
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

	//var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::;
	db, _, er := DatabaseInit("main", "")
	if er != nil {
		return er
	}
	defer a.closeMyDb(db)
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

	errAsset := GetJson("https://rahavard365.com/api/search/items?type=asset", &rawsAsset, &a.HttpLock)

	if errAsset != nil {
		fmt.Printf("error -> getjson() -> %v \n",errAsset)
		return errAsset
	}
	errIndex := GetJson("https://rahavard365.com/api/search/items?type=index", &rawsIndex, &a.HttpLock)

	if errIndex != nil {
		fmt.Printf("error -> getjson() -> %v \n",errAsset)
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

/*out stock */
func (a StockProvider) OutStockList(dbLock *sync.Mutex) error {

	//var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::;
	db, _, er := DatabaseInit("main", "")
	if er != nil {
		return er
	}
	defer a.closeMyDb(db)
	items, err := GetNemadList(db, dbLock)
	if err != nil {
		return err
	}
	var data = [][]string{{}}
	for _, k := range items {
		//fmt.Printf("%v %v %v\n", k.EntityType, k.EntityId, k.TradeSymbol)
		data = append(data, []string{k.EntityType, strconv.FormatInt(k.EntityId, 10), k.TradeSymbol})
	}
	//:::::::: write to csv
	var s string = path.Join(GetRootCache(), "stock_list.csv")
	file, err := os.Create(s)
	if err != nil {
		return errors.New(fmt.Sprintf("OutStockList -> Cannot create file %s", err))
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		if err != nil {
			return errors.New(fmt.Sprintf("OutStockList -> Cannot create file %s", err))
		}
	}
	fmt.Printf("has been successfully created : %s \n", s)
	return nil
}
func (a StockProvider) OutTemWatchList(dbLock *sync.Mutex) error {

	//var fullPath string
	//:::::::::::::::::::::::::::::::::::::::::;
	db, _, er := DatabaseInit("main", "")
	if er != nil {
		return er
	}
	defer a.closeMyDb(db)
	items, err := GetNemadList(db, dbLock)
	if err != nil {
		return err
	}
	data := WatchListItem{}
	t := []WatchStock{}
	t1 := true
	for _, k := range items {
		//fmt.Printf("%v %v %v\n", k.EntityType, k.EntityId, k.TradeSymbol)

		g := WatchStock{}
		g.TimeFrame = []string{"d"}
		g.NameEn = strconv.FormatInt(k.EntityId, 10)
		g.AssetCode = strconv.FormatInt(k.EntityId, 10)
		g.IsIndex = true
		if strings.Contains(k.EntityType, "index") {
			g.IsIndex = false
		}

		g.IsAdj = &t1
		t = append(t, g)
	}
	data.Tehran = t
	data.Crypto = []WatchStock{}

	//:::::::: write to csv
	var s string = path.Join(GetRootCache(), "temp_watch_list.csv")
	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(s, file, 0644)

	fmt.Printf("has been successfully created : %s \n", s)
	return nil
}

func (a StockProvider) AddStockToWatchList(provider EProvider, stockName string, stockCode string, index_t bool, adj *bool) error {
	w, e := a.ReadJsonWatchList()

	if e != nil {
		return errors.New(fmt.Sprintf("config read failed [%v] ", e))
	}
	g := WatchStock{}
	if provider == Avard {
		for _, l := range w.Tehran {
		if l.AssetCode == stockCode{
			fmt.Printf("stock is exist %v %v \n",stockCode,stockName)
			return nil
		}
		}

		g.TimeFrame = []string{"d"}
		g.NameEn = stockName
		g.AssetCode = stockCode
		g.IsIndex = index_t
		g.IsAdj = adj

		w.Tehran = append(w.Tehran, g)
	}
	if provider == Binance {
		for _, l := range w.Crypto {
			if l.AssetCode == stockCode{
				fmt.Printf("stock is exist %v %v \n",stockCode,stockName)
				return nil
			}
		}

		g.TimeFrame = []string{"d"}
		g.NameEn = stockName
		g.AssetCode = stockCode
		g.IsIndex = index_t
		g.IsAdj = adj

		w.Crypto = append(w.Crypto, g)
	}

	//+++++++++++++++++++++++++

	s := path.Join(GetRootCache(), "watchList.json")
	if err := os.Remove(s); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to remove  file for %v", err)
	}

	file, _ := json.MarshalIndent(w, "", " ")
	_ = ioutil.WriteFile(s, file, 0644)
	return nil
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func (a StockProvider) isHasTimeFrame(timeframe ETimeFrame, stock WatchStock) bool {

	if stock.TimeFrame == nil {
		return true
	}
	if len(stock.TimeFrame) == 0 {
		return true
	}

	for _, g := range stock.TimeFrame {
		if strings.ToLower(timeframe.ToString2()) == strings.ToLower(g) {
			return true
		}
	}
	return false
}
func (a StockProvider) isHasAdjust(stock WatchStock) bool {

	if stock.IsAdj == nil {
		return true
	}

	return *stock.IsAdj
}

func (a StockProvider) avardAssetProcess(parentWaitGroup *sync.WaitGroup, readFromLast bool, watchStock WatchStock) error {

	a.sem.Acquire(context.Background(), 1)
	defer a.sem.Release(1)
	defer parentWaitGroup.Done()
	if watchStock.NameEn == "" || watchStock.AssetCode == "" {
		//parentWaitGroup.Done()
		return errors.New("field is empty ")

	}
	var databaseLock sync.Mutex
	var wg sync.WaitGroup
	if a.Provider == Avard {

		/*if watchStock.IsIndex == true {
			wg.Add(2)
		} else {
			wg.Add(8)
		}*/
		var num_d1 time.Duration = 15000 //1979-01-12
		var num_h4 time.Duration = 1000
		var num_h2 time.Duration = 500
		var num_h1 time.Duration = 500

		if a.isHasTimeFrame(H1, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h1), EndTime: time.Now(), TimeFrame: H1, TypeChart: Normal})
		}
		if a.isHasTimeFrame(D1, watchStock) {

			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_d1), EndTime: time.Now(), TimeFrame: D1, TypeChart: Normal})
		}

		if watchStock.IsIndex == false {
			if a.isHasTimeFrame(H2, watchStock) {

				a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h2), EndTime: time.Now(), TimeFrame: H2, TypeChart: Normal})
			}
			if a.isHasTimeFrame(H4, watchStock) {

				a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h4), EndTime: time.Now(), TimeFrame: H4, TypeChart: Normal})
			}
			if a.isHasAdjust(watchStock) {
				if a.isHasTimeFrame(H1, watchStock) {
					a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h1), EndTime: time.Now(), TimeFrame: H1, TypeChart: Adj})
				}
				if a.isHasTimeFrame(H2, watchStock) {
					a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h2), EndTime: time.Now(), TimeFrame: H2, TypeChart: Adj})
				}
				if a.isHasTimeFrame(H4, watchStock) {
					a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_h4), EndTime: time.Now(), TimeFrame: H4, TypeChart: Adj})
				}
				if a.isHasTimeFrame(D1, watchStock) {
					a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * num_d1), EndTime: time.Now(), TimeFrame: D1, TypeChart: Adj})
				}
			}
		}
	} else if a.Provider == Binance {

		if a.isHasTimeFrame(M15, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: M15, TypeChart: Normal})
		}
		if a.isHasTimeFrame(H1, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H1, TypeChart: Normal})
		}
		if a.isHasTimeFrame(H2, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * 250), EndTime: time.Now(), TimeFrame: H2, TypeChart: Normal})
		}
		if a.isHasTimeFrame(H4, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * 360), EndTime: time.Now(), TimeFrame: H4, TypeChart: Normal})
		}
		if a.isHasTimeFrame(D1, watchStock) {
			a.procMake(StockQuery{WaitGroupobj: &wg, DBLock: &databaseLock, ReadfromLast: readFromLast, Stock: watchStock, Duration: -time.Duration(time.Hour * 24 * 9000), EndTime: time.Now(), TimeFrame: D1, TypeChart: Normal})
		}
	} else {
		return errors.New("not selected :( ")
	}
	wg.Wait()

	return nil
}
func (a StockProvider) Run(readfromLast bool, isSeq bool, timer_minute *int64) error {
	var e error

	a.sem = semaphore.NewWeighted(10)
	a._WatchListItem, e = a.ReadJsonWatchList()
	a.IsSeqRunProcess = isSeq
	if e != nil {
		return errors.New(fmt.Sprintf("config read failed [%v] ", e))
	}
	if timer_minute == nil {
		//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
		var wg sync.WaitGroup
		if a.Provider == Avard {
			wg.Add(len(a._WatchListItem.Tehran))
			for _, g := range a._WatchListItem.Tehran {
				if a.IsSeqRunProcess {
					a.avardAssetProcess(&wg, readfromLast, g)
				} else {
					go a.avardAssetProcess(&wg, readfromLast, g)
				}

			}
		} else if a.Provider == Binance {
			wg.Add(len(a._WatchListItem.Crypto))
			for _, g := range a._WatchListItem.Crypto {
				if a.IsSeqRunProcess {
					a.avardAssetProcess(&wg, readfromLast, g)
				} else {
					go a.avardAssetProcess(&wg, readfromLast, g)
				}

			}
		}
		wg.Wait()
		//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
	} else {
		tick := time.Tick(time.Duration(*timer_minute) * time.Minute)

		isWorking := false
		for {

			if isWorking {
				continue
			}

			select {
			case <-tick:
				fmt.Printf("timer -> %v\n", time.Now().Format(time.ANSIC))
				isWorking = true
				//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
				var wg sync.WaitGroup
				if a.Provider == Avard {
					wg.Add(len(a._WatchListItem.Tehran))
					for _, g := range a._WatchListItem.Tehran {
						if a.IsSeqRunProcess {
							a.avardAssetProcess(&wg, readfromLast, g)
						} else {
							go a.avardAssetProcess(&wg, readfromLast, g)
						}

					}
				} else if a.Provider == Binance {
					wg.Add(len(a._WatchListItem.Crypto))
					for _, g := range a._WatchListItem.Crypto {
						if a.IsSeqRunProcess {
							a.avardAssetProcess(&wg, readfromLast, g)
						} else {
							go a.avardAssetProcess(&wg, readfromLast, g)
						}

					}
				}
				wg.Wait()
				//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
				isWorking = false
			}
		}

	}
	return nil
}
