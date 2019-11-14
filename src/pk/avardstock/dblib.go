package avardstock

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	h "github.com/qasemt/helper"
	"sync"
)

var pathdb string

type Nemad struct {
	ID        string
	GroupCode string
	GroupName string
	NemadCode string
	NameEn    string
	NameFa    string
	NameFull  string
	AvardCode int
}

/*type StockFromWebService struct {
	ID        uint64  `gorm:"primary_key; not null"`
	AssetId   string  `gorm:"unique_index:indexok;not null"`
	TimeFrame int     `gorm:"unique_index:indexok; not null"`
	TypeChart int     `gorm:"unique_index:indexok; not null;gorm:"default:0"`
	Time      float64 `gorm:"unique_index:indexok;not null" json:"time"`
	O         float64 `json:"open"`
	H         float64 `json:"high"`
	L         float64 `json:"low"`
	C         float64 `json:"close"`
	V         float64 `json:"volume"`
}*/
type StockFromWebService struct {
	ID        uint64  `gorm:"primary_key; not null"`
	AssetId   string  `gorm:"not null"`
	TimeFrame int     `gorm:"not null"`
	TypeChart int     `gorm:"not null;gorm:"default:0"`
	Time      int64   `gorm:"not null" json:"time"`
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
func dbSync(d *gorm.DB)  error {
	if d == nil {
		return errors.New("db not init")
	}
	// Migrate the schema
	(d).AutoMigrate(&Nemad{})
	(d).AutoMigrate(&StockFromWebService{})
	//db.Delete(&StockFromWebService{}, "")
	return nil
}
func DatabaseInit(dbName string ,db1 *gorm.DB) (*gorm.DB, error ){
	var err error
	if db1 == nil {
		db1, err = gorm.Open("sqlite3", dbName)
		if err != nil {
			db1 = nil
			panic("failed to connect database")
		}
	}

	dbSync(db1)

	return db1, nil
}
func InsertStocks(d *gorm.DB,k *sync.Mutex,stockList []StockFromWebService, assetid string, timeframe h.ETimeFrame, tc h.ETypeChart) error {
defer k.Unlock()
	k.Lock()
	if d == nil {
		return errors.New("db not init")
	}
	for i := 0; i < len(stockList); i++ {
		t := stockList[i]
		t.AssetId = assetid
		t.TimeFrame = int(timeframe)
		t.TypeChart = int(tc)

		if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=? and time  = ?",assetid,timeframe,int(tc),t.Time).Order("time desc").Limit(1).First(&t).Error; gorm.IsRecordNotFoundError(e1) {
			d.Create(&t)
		}

	}
	return nil
}
func getLastRecord(d *gorm.DB,assetid string, timeframe int, tc h.ETypeChart, out *StockFromWebService) error {

	if d == nil {
		return errors.New("db not init")
	}

	if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=?",assetid,timeframe,int(tc)).Order("time desc").Limit(1).First(out).Error; gorm.IsRecordNotFoundError(e1) {
	//fmt.Println("getLastRecord () -> record not found")
	}

	if d.Error != nil {
		return d.Error
	}
	return nil
}
func getRecordes(d *gorm.DB,assetid string, timeframe h.ETimeFrame, tc h.ETypeChart) ([]StockFromWebService ,error) {

	if d == nil {
		return nil,errors.New("db not init")
	}
var items  []StockFromWebService

	if e1 := d.Where("asset_id = ? and time_frame = ? and type_chart=?",assetid,int(timeframe),int(tc)).Order("time").Find(&items).Error; gorm.IsRecordNotFoundError(e1) {
		//fmt.Println("getLastRecord () -> record not found")
	}

	if d.Error != nil {
		return nil,d.Error
	}
	return items,nil
}