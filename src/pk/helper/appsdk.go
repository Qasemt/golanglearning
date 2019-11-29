package helper

import (
	"fmt"
	"time"
)
//::::::::::::::::::::::::::::: STRUCT
type StockItem struct {
	D string
	T string

	O float64
	H float64
	L float64
	C float64
	V float64

	BV float64
}

type TimeRange struct {
	File_name string
	Begin     time.Time
	End       time.Time
}
type WatchStock struct {
	AssetCode string `json:"asset_code"`
	NameEn    string `json:"nameEn"`
	IsIndex   bool   `json:"is_index"`
}
type WatchListItem struct {
	Apikey string
	Secret string
	Tehran []WatchStock
	Crypto []WatchStock
}
//::::::::::::::::::::::::::::: ENUM
type ETimeFrame int
type ETypeChart int
type EProvider int

const (
	Binance EProvider =1
	Avard   EProvider =2
)
const (
	M1  ETimeFrame = 1
	M15 ETimeFrame = 15
	H1  ETimeFrame = 60
	H2  ETimeFrame = 120
	H4  ETimeFrame = 240
	D1  ETimeFrame = 1440
)
const (
	Normal ETypeChart = 0
	Adj    ETypeChart = 1
)
//::::::::::::::::::::::::::::: ENUM TO STRING

func (e ETypeChart) ToTypeChartStr() string {
	switch e {
	case Adj:
		return "Adjust"
	case Normal:
		return "Normal"
	}
	return ""
}
func (e ETypeChart) ToString() string {
	switch e {
	case Normal:
		return ""
	case Adj:
		return "1"

	default:
		return ""
	}
}


func (e ETimeFrame) ToString2() string {
	switch e {
	case M1:
		return "1min"
	case M15:
		return "15min"
	case H1:
		return "1h"
	case H2:
		return "2h"
	case H4:
		return "4h"
	case D1:
		return "D"

	default:
		return fmt.Sprintf("%d", int(e))
	}
}
func (e ETimeFrame) ToString() string {
	switch e {
	case M1:
		return "1m"
	case M15:
		return "15m"
	case H1:
		return "1h"
	case H2:
		return "2h"
	case H4:
		return "4h"
	case D1:
		return "1d"

	default:
		return fmt.Sprintf("%d", int(e))
	}
}
func (e ETimeFrame) ToMinuth() int {
	switch e {
	case M1:
		return 1
	case M15:
		return 15
	case H1:
		return 60
	case H2:
		return 120
	case H4:
		return 240
	case D1:
		return 1440

	}
	return 0
}


