package binance

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"os"
	"path"
	"strconv"
	"time"


)



var cachePath = "d:/cache2/"
var cacheLastCandel = "lastcandel"


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
func int64FromString(raw interface{}) (int64, error) {
	str, ok := raw.(string)
	if !ok {
		return 0, nil
	}

	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, nil
	}
	return n, nil
}
func timeFromUnixTimestampFloat(raw interface{}) (time.Time, error) {
	ts, ok := raw.(float64)
	if !ok {
		return time.Time{}, nil
	}
	return time.Unix(0, int64(ts)*int64(time.Millisecond)), nil
}




//----------------------------------------------- NEW METHOD
var lastCandelTime time.Time

func GetDateRange(duration time.Duration, end time.Time) []TimeRange {
	day_rang := []TimeRange{}
	start := end.Add(-duration)
	diff := end.Sub(start) / (24 * time.Hour)
	for i := 1; i <= int(diff); i++ {
		var tt = start.AddDate(0, 0, i)
		var d1 TimeRange
		d1.File_name = TimeToString(tt, "yyyymmdd") + ".csv"
		y, m, d := tt.Date()
		d1.Begin = time.Date(y, m, d, 0, 0, 0, 0, tt.Location())
		d1.End = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())
		day_rang = append(day_rang, d1)
	}
	return day_rang
}
func DownloadAsset(assetName string, item TimeRange, timefram ETimeFrame) ([]StockItem, error) {
	var _rawKlines = [][]interface{}{}
	fmt.Println(item)
	cacheFolderPath := path.Join(cachePath, assetName, (timefram).ToString())
	var itemsFinal []StockItem
	startStr := strconv.FormatInt(UnixMilli(item.Begin), 10)
	endStr := strconv.FormatInt(UnixMilli(item.End), 10)
	err := GetJson("https://api.binance.com/api/v3/klines?symbol="+assetName+"&interval="+timefram.ToString()+"&startTime="+startStr+"&endTime="+endStr, &_rawKlines)

	if err != nil {
		return nil, err
	}
	if _rawKlines == nil {
		return nil, errors.New("DownloadAsset failed ... binance block link")
	}

	for _, k := range _rawKlines {
		var v StockItem
		time1, _ := timeFromUnixTimestampFloat(k[0])
		lastCandelTime = time1
		v.D = UnixTimeStrToFormatDT(time1, true)
		v.T = UnixTimeStrToFormatDT(time1, false)

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

	if len(itemsFinal) > 0 {
		if !OutToCSVFile(itemsFinal, cacheFolderPath, item.File_name, false) {
			return nil, errors.New("outtocsvfile : failed :(")
		}
	}
	return itemsFinal, nil
}
func MakeCacheBase15M(assetName string, duration time.Duration, end time.Time) error {

	dayRang := GetDateRange(duration, end)
	var last_File_date_ProcessinM15 string

	//::::::::::::::::::::::::::::::
	for i := 0; i < len(dayRang); i++ {
		it := dayRang[i]
		file_15m := path.Join(cachePath, assetName, (M15).ToString(), it.File_name)
		//________________________________________________
		if IsExist(file_15m) && CompareDate(time.Now(), it.Begin) {
			last_File_date_ProcessinM15 = TimeToString(it.Begin, "yyyymmdd") + ".csv"
			var err = os.Remove(file_15m)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
		//_________________________________________________
		if !IsExist(file_15m) {
			_, e := DownloadAsset(assetName, it, M15)
			if e != nil {
				return e
			}
		}
		//_________________________________________________ make last candel 15m from 1m
		var candel1m TimeRange
		if CompareDate(time.Now(), it.Begin) {

			candel1m.End = time.Now()
			candel1m.Begin = lastCandelTime.Add(time.Minute * 1)
			candel1m.File_name = "lastcandels" + ".csv"
			lastCandels1m, e := DownloadAsset(assetName, candel1m, M1)
			if e != nil {
				return e
			}
			filePath15 := path.Join(cachePath, assetName, (M15).ToString(), last_File_date_ProcessinM15)
			res := makeLastCandelFromTimeFrame1M(filePath15, lastCandels1m)
			if res != nil {
				return errors.New("join last candel from 1m time frame to 15m file has been failed.")
			}
		}
	}
	return nil
}
func MakeCandel(arr [][]string,  totimeframe ETimeFrame) ([]string, error) {
	//	20191108,191500,8762.2200,8786.8200,8756.8200,8785.1600,313.1609,0.0000
	l := make([]string, 8)
	if totimeframe == H1 {
		l[0] = arr[3][0] //date
		l[1] = arr[3][1] //time
		l[2] = arr[0][2] //O
		l[3] = arr[0][3] //H

		l[4] = arr[3][4] //L
		l[5] = arr[3][5] //C

		var v float64
		for i := 0; i < len(arr); i++ {
			v = v + ToFloat(arr[i][6])
		}

		l[6] = strconv.FormatFloat(v,'f', 4, 64)

		l[7] = arr[0][7] //i
	}
	return l, nil
}
func MakeCacheHourly(assetName string, baseTimeFrame ETimeFrame, totimeframe ETimeFrame, duration time.Duration, end time.Time) error {
	//
	dayRang := GetDateRange(duration, end)

	for i := 0; i < len(dayRang); i++ {
		it := dayRang[i]
		file_15m := path.Join(cachePath, assetName, (baseTimeFrame).ToString(), it.File_name)

		if !IsExist(file_15m) {
			return errors.New(fmt.Sprintln("MakeCacheHourly: file not found", file_15m))
		}
		res, list := GetJsonToArry(file_15m)
		if !res {
			return errors.New(fmt.Sprintln("MakeCacheHourly:read failed", file_15m))
		}
		var recs [][]string
		c:=0
		for i := 0; i < len(list); i++ {

			if c<4 {
				c ++
				recs = append(recs, list[i])
			}else if c ==4 {
				t, _ := MakeCandel(recs, H1)
				c=0
				fmt.Println(t)
				recs =recs[:0]
			}

		}
		println(list)
	}

	return nil
}

func makeLastCandelFromTimeFrame1M(mainFilePath string, listLastCandel1m []StockItem) error {
	if !IsExist(mainFilePath) {
		return errors.New(fmt.Sprintln("file not found ", mainFilePath))
	}
	var final_array []StockItem
	if len(listLastCandel1m) > 0 {
		var f = listLastCandel1m[0]
		var l = listLastCandel1m[len(listLastCandel1m)-1]
		var candel15m StockItem
		candel15m.D = l.D
		candel15m.T = l.T

		candel15m.O = f.O
		candel15m.H = f.H

		candel15m.L = l.L
		candel15m.C = l.C

		for _, it := range listLastCandel1m {
			candel15m.V = candel15m.V + it.V
		}

		final_array = append(final_array, candel15m)
	}

	if !AppendToCSVFile(final_array, mainFilePath, false) {
		return errors.New(fmt.Sprintln("make laset candel from 1m failed "))
	}

	return nil
}
