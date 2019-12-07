package avardstock

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	h "github.com/qasemt/helper"
	"os"
	"path"
	"strings"
	"sync"
)

var pathdb string

//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::STRUCT INIT
type Nemad struct {
	ID        uint64 `gorm:"primary_key; not null"`
	NemadID   string
	GroupCode string
	GroupName string
	NemadCode string
	NameEn    string
	NameFa    string
	NameFull  string
	AvardCode int
}

type NemadAvard struct {
	ID          uint64 `gorm:"primary_key; not null"`
	TypeId      string `json:"type_id"`
	Type        string `json:"type"`
	EntityId    int64  `json:"entity_id"`
	EntityType  string `json:"entity_type"`
	TradeSymbol string `json:"trade_symbol"`
	Title       string `json:"title"`
}

type StockFromWebService struct {
	ID        uint64  `gorm:"primary_key; not null"`
	AssetId   string  `gorm:"unique_index:indexok;not null"`
	IsIndex   int     `gorm:"not null;default:0"`
	TimeFrame int     `gorm:"unique_index:indexok; not null"`
	TypeChart int     `gorm:"unique_index:indexok; not null;default:0"`
	Time      int64   `gorm:"unique_index:indexok;not null" json:"time"`
	O         float64 `json:"open"`
	H         float64 `json:"high"`
	L         float64 `json:"low"`
	C         float64 `json:"close"`
	V         float64 `json:"volume"`
}

func (e StockFromWebService) TOString() string {

	return fmt.Sprintf("ID :%v Time:%f ,frame: %v type :%v", e.ID, e.Time, e.TimeFrame, e.TypeChart)
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func DatabaseInit(dbName1 string, timefrm string, db1 *gorm.DB) (*gorm.DB, string, error) {
	var err error
	var dirdbstr string
	var fullPath string

	dirdbstr = path.Join(h.GetRootCache(), "cache")

	if _, err := os.Stat(dirdbstr); os.IsNotExist(err) {
		os.MkdirAll(dirdbstr, os.ModePerm)
	}

	if timefrm == "" {
		fullPath = path.Join(dirdbstr, fmt.Sprintf("%v.bin", dbName1))
	} else {
		//fullPath = path.Join(dirdbstr, fmt.Sprintf("%v_%v.bin", dbName1, timefrm))
		fullPath = path.Join(dirdbstr, fmt.Sprintf("%v.bin", dbName1))
	}

	if db1 == nil {
		db1, err = gorm.Open("sqlite3", fullPath)
		if err != nil {
			db1 = nil
			panic("failed to connect database")
		}
	}

	return db1, fullPath, nil
}

func Migrate(dbName1 string, isp *StockProvider) error {

	var fullPath string
	var db *gorm.DB
	defer isp.closeMyDb(db)
	db, fullPath, er := DatabaseInit(dbName1, "", db)

	if er != nil {
		return errors.New(fmt.Sprintf("err:%v %v", er, fullPath))
	}

	if strings.Contains(fullPath, "main.bin") {
		db.AutoMigrate(&Nemad{})
	} else {
		db.AutoMigrate(&StockFromWebService{})
	}
	return nil
}
func InsertStocks(d *gorm.DB, k *sync.Mutex, isIndex bool, stockList []StockFromWebService, assetid string, timeframe h.ETimeFrame, tc h.ETypeChart) error {
	defer k.Unlock()

	k.Lock()
	if d == nil {
		return errors.New("db not init")
	}
	/*for i := 0; i < len(stockList); i++ {
		t := stockList[i]
		t.AssetId = assetid
		t.TimeFrame = int(timeframe)
		t.TypeChart = int(tc)
		if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=? and time  = ?", assetid, timeframe, int(tc), t.Time).Order("time desc").Limit(1).First(&t).Error; gorm.IsRecordNotFoundError(e1) {
			d.Create(&t)
		}
	}*/
	//________________
	valueStrings := []string{}
	for _, t := range stockList {

		t.AssetId = assetid
		t.TimeFrame = int(timeframe)
		t.TypeChart = int(tc)
		if isIndex == true {
			t.IsIndex = 1
		}

		valueStrings = append(valueStrings, fmt.Sprintf("(\"%s\", %v,%v, %v, %v, %v, %v, %v, %v, %v)", t.AssetId, t.IsIndex, t.TimeFrame, t.TypeChart, t.Time, t.O, t.H, t.L, t.C, t.V))
	}

	smt := `INSERT OR REPLACE  INTO stock_from_web_services (asset_id,is_index,time_frame,type_chart,TIME,o,h,l,c,v) VALUES   %s ;`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	tx := d.Begin()
	if err := tx.Exec(smt).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
func getLastRecord(d *gorm.DB, k *sync.Mutex, assetid string, timeframe int, tc h.ETypeChart, out *StockFromWebService) error {
	defer k.Unlock()
	k.Lock()
	if d == nil {
		return errors.New("db not init")
	}

	if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=?", assetid, timeframe, int(tc)).Order("time desc").Limit(1).First(out).Error; gorm.IsRecordNotFoundError(e1) {
		//fmt.Println("getLastRecord () -> record not found")
	}

	if d.Error != nil {
		return d.Error
	}
	return nil
}
func getRecordesStock(d *gorm.DB, k *sync.Mutex, assetid string, timeframe h.ETimeFrame, tc h.ETypeChart) ([]StockFromWebService, error) {
	defer k.Unlock()
	k.Lock()

	if d == nil {
		return nil, errors.New("db not init")
	}
	var items []StockFromWebService

	if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=?", assetid, int(timeframe), int(tc)).Order("time").Find(&items).Error; gorm.IsRecordNotFoundError(e1) {
		//fmt.Println("getLastRecord () -> record not found")
	}

	if d.Error != nil {
		return nil, d.Error
	}
	return items, nil
}
func GetNemadList(d *gorm.DB, k *sync.Mutex) ([]NemadAvard, error) {
	defer k.Unlock()
	k.Lock()

	if d == nil {
		return nil, errors.New("db not init")
	}
	var items []NemadAvard
	if e1 := d.Find(&items).Error; gorm.IsRecordNotFoundError(e1) {
		fmt.Println("getNemadList () -> record not found")
	}
	if d.Error != nil {
		return nil, d.Error
	}
	return items, nil
}

func InsertAssetInfoFromAvard(d *gorm.DB, k *sync.Mutex, avardsAsset []NemadAvard) error {
	defer k.Unlock()
	k.Lock()
	if d == nil {
		return errors.New("db not init")
	}
	d.DropTableIfExists(&NemadAvard{})
	d.AutoMigrate(&NemadAvard{})

	//-----------------------------------------------------
	valueStrings := []string{}

	for _, f := range avardsAsset {
		valueStrings = append(valueStrings, fmt.Sprintf("(\"%s\", \"%s\", %v,\"%s\",\"%s\",\"%s\")", f.TypeId, f.Type, f.EntityId, f.EntityType, f.TradeSymbol, f.Title))
	}

	smt := `INSERT INTO nemad_avards (type_id,type,entity_id,entity_type,trade_symbol,title) VALUES  %s ;`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	tx := d.Begin()
	if err := tx.Exec(smt).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
