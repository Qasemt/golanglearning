package avardstock

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"strconv"
	"time"
)

var cachePath = "d:/cache2/tehran/"

func downloadAsset(assetName string, item TimeRange, timefram ETimeFrame) ([]StockFromWebService, error) {
	var _rawKlines = []StockFromWebService{}
	fmt.Println(item)

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
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange.asset%3A66%3Areal_close%3Atype1&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &_rawKlines)

	if err != nil {
		return nil, err
	}

	if _rawKlines == nil {
		return nil, errors.New("downloadAsset failed ... binance block link")
	}

	return _rawKlines, nil
}

func Make(assetName string, duration time.Duration, end time.Time, timeFrame ETimeFrame) error {

	var it = TimeRange{}
	//var itemsFinal []StockItem
	it.Begin = end.Add(duration)
	it.End = end
	itemsRaws, e := downloadAsset(assetName, it,timeFrame )

	if e != nil {
		return e
	}

	db, er := DatabaseInit()
	if er != nil {
		return er
	}
	
	if len(itemsRaws) > 0 {

		defer db.Close()

		err := InsertStocks(itemsRaws, assetName,timeFrame , db)

		if err != nil {
			return err
		}
	}

	/*for _, k := range _rawKlines {
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

	}*/

	return nil
}
