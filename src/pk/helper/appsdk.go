package helper

import (
	"fmt"
	"time"
)

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
type StockFromWebService struct {
	AssetId string  `json:"asset_id"`
	Time    float64  `json:"time"`
	O       float64 `json:"open"`
	H       float64 `json:"high"`
	L       float64 `json:"low"`
	C       float64 `json:"close"`
	V       float64`json:"volume"`
}

//:::::::::::::::::::::::::::::
type ETimeFrame int

const (
	M1  ETimeFrame = 1
	M15 ETimeFrame = 2
	H1  ETimeFrame = 3
	H2  ETimeFrame = 4
	H4  ETimeFrame = 5
	D1  ETimeFrame = 6
)

type TimeRange struct {
	File_name string
	Begin     time.Time
	End       time.Time
}

func (e ETimeFrame) String() string {
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

//:::::::::::::::::::::::::::::
