package avardstock

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	h "github.com/qasemt/helper"
)

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
type StockFromWebService struct {
	ID        uint64  `gorm:"primary_key; not null"`
	AssetId   string  `gorm:"index:tt;not null"`
	TimeFrame int     `gorm:"index:tt; not null"`
	Time      float64 `gorm:"index:tt;not null" json:"time"`
	O         float64 `json:"open"`
	H         float64 `json:"high"`
	L         float64 `json:"low"`
	C         float64 `json:"close"`
	V         float64 `json:"volume"`
}

func DatabaseInit() (*gorm.DB, error) {

	db, err := gorm.Open("sqlite3", "./stock.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Nemad{})
	db.AutoMigrate(&StockFromWebService{})
	db.Delete(&StockFromWebService{}, "")
	return db, nil
}
func InsertStocks(stockList []StockFromWebService, assetid string, timeframe h.ETimeFrame, db *gorm.DB) (bool, error) {
	var err error
	db, err = gorm.Open("sqlite3", "./stock.db")
	if err != nil {
		panic("failed to connect database")
	}

	for i := 0; i < len(stockList); i++ {
		t := stockList[i]
		t.AssetId = assetid
		t.TimeFrame = int(timeframe)
		//	if db.NewRecord(t) {
		db.Create(&t)
		//	}
	}
	return true, nil
}
