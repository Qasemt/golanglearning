package avardstock

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"


	"os"
	"path"
	"strconv"
	"time"
)

var cachePath = "d:/cache2/tehran/"

func getDateRange(duration time.Duration, end time.Time,frame ETimeFrame ) []TimeRange {
	day_rang := []TimeRange{}
	start := end.Add(duration)
	var diff time.Duration
	if frame ==D1 {
		diff = end.Sub(start) /  ((24 * time.Hour)*360)
		diff=diff+1
	}
	for i := 1; i <= int(diff); i++ {
		var tt = start.AddDate(0, 0, i)
		var d1 TimeRange

		y, m, d := tt.Date()

		if frame == D1 {
			d1.File_name =	 fmt.Sprintf("%4d%02d%02d",tt.Year(), 1, 1)+ ".csv"
			d1.Begin = time.Date(y, 1,1, 0, 0, 0, 0, tt.Location())
			d1.End = time.Date(y, 12,31, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())
		}else {
			d1.File_name = TimeToString(tt, "yyyymmdd") + ".csv"
			d1.Begin = time.Date(y, m, d, 0, 0, 0, 0, tt.Location())
			d1.End = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())
		}
		day_rang = append(day_rang, d1)
	}
	return day_rang
}
func downloadAsset(assetName string, item TimeRange, timefram ETimeFrame) ([]StockItem, error) {
	var _rawKlines = []StockFromWebService{}
	fmt.Println(item)
	cacheFolderPath := path.Join(cachePath, assetName, (timefram).String())
	var itemsFinal []StockItem
	startStr := strconv.FormatInt(item.Begin.Unix(), 10)
	endStr := strconv.FormatInt(item.End.Unix(), 10)

	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange.asset%3A66%3Areal_close%3Atype1&resolution=d&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &_rawKlines)

	if err != nil {
		return nil, err
	}
	if _rawKlines == nil {
		return nil, errors.New("downloadAsset failed ... binance block link")
	}

	for _, k := range _rawKlines {
		var v StockItem
		time1 :=time.Unix(0, int64(k.Time) * int64(time.Millisecond))
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
		if !OutToCSVFile(itemsFinal, cacheFolderPath, item.File_name, false) {
			return nil, errors.New("outtocsvfile : failed :(")
		}
	}
	return itemsFinal, nil
}

func Make(assetName string, duration time.Duration, end time.Time, timefram ETimeFrame)   error {

	dayRang := getDateRange(duration, end,timefram)

	for i := 0; i < len(dayRang); i++ {
		it := dayRang[i]
		fmt.Println(it)
		file_path:= path.Join(cachePath, assetName, (timefram).String(), it.File_name)
		if IsExist(file_path) && it.Begin.Year() == time.Now().Year() && it.Begin.Month() == time.Now().Month() && it.Begin.Day() == time.Now().Day() {
			var err = os.Remove(file_path)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
		//---------
		if IsExist(file_path) {
			continue
		}

		_, e := downloadAsset(assetName,it,timefram)
		if e != nil {
			return e
		}
	}
	return  nil
}
