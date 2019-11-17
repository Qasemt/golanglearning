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

//:::::::::::::::::::::::::::::
type ETimeFrame int
type ETypeChart int

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

type TimeRange struct {
	File_name string
	Begin     time.Time
	End       time.Time
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

//:::::::::::::::::::::::::::::
