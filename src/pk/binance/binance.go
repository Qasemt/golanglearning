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



func GetAsset(assetName string, start time.Time, end time.Time, timeframe ETimeFrame, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(cachePath, assetName, timeframe.String())
	last_candel := path.Join(cachePath, cacheLastCandel, assetName, (timeframe).String()+".csv")
	var items_final []StockItem

	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)
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
	for _, item := range day_rang {
		items_final = items_final[:0]

		start_str := strconv.FormatInt(UnixMilli(item.Begin), 10)
		end_str := strconv.FormatInt(UnixMilli(item.End), 10)

		var final_out = path.Join(dir_cache_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if IsExist(final_out) && item.Begin.Year() == time.Now().Year() && item.Begin.Month() == time.Now().Month() && item.Begin.Day() == time.Now().Day() {
			var err = os.Remove(final_out)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		if IsExist(final_out) {
			continue
		}

		fmt.Printf(">>> %v %v %v \n", assetName, timeframe.String(), TimeToString(item.Begin, "yyyy-mm-dd"))
		err := GetJson("https://api.binance.com/api/v3/klines?symbol="+assetName+"&interval="+timeframe.String()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
		if err != nil {
			return err
		}

		for _, k := range rawKlines {
			var v StockItem
			time1, _ := timeFromUnixTimestampFloat(k[0])

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

			items_final = append(items_final, v)

		}
		if len(items_final) > 0 {
			if !OutToCSVFile(items_final, dir_cache_path, item.File_name, false) {
				return errors.New("get asset >>> out to csv failed")
			}
		}
	}
	if !JoinCSVFiles(dir_cache_path, file_list, final_out, last_candel) {
		return errors.New("get asset >>> join last candel failed")
	}

	return nil
}
func GetAssetCreateLastCandel(assetName string, end time.Time, timeframe ETimeFrame) error {

	rawKlines := [][]interface{}{}
	var dir_cache_path string
	dir_cache_path = path.Join(cachePath, cacheLastCandel, assetName)

	var items_final []StockItem

	var item TimeRange
	item.File_name = timeframe.String() + ".csv"
	var diff time.Duration
	if timeframe == H1 {
		diff = time.Hour * 1
	} else if timeframe == H2 {
		diff = time.Hour * 2
	} else if timeframe == H4 {
		diff = time.Hour * 4
	} else if timeframe == M15 {
		diff = time.Minute * 15
	}

	if timeframe == D1 {
		y, m, d := end.Date()

		item.Begin = time.Date(y, m, d, 0, 0, 0, 0, end.Location())
		item.End = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), end.Location())
	} else {
		item.Begin = end.Add(-diff)
		item.End = end
	}

	items_final = items_final[:0]
	//fmt.Println(item)
	start_str := strconv.FormatInt(UnixMilli(item.Begin), 10)
	end_str := strconv.FormatInt(UnixMilli(item.End), 10)

	err := GetJson("https://api.binance.com/api/v3/klines?symbol="+assetName+"&interval=1m&startTime="+start_str+"&endTime="+end_str, &rawKlines)
	if err != nil {
		return err
	}
	if len(rawKlines) == 0 {
		e := fmt.Errorf("GetAssetCreateLastCandel: failed plz change your ip ,current ip block from binance -> %v time-frame :%v", assetName, timeframe)
		return e
	}
	fmt.Println(":::", assetName, timeframe, TimeToString(item.Begin, ""), TimeToString(item.End, ""), len(rawKlines))
	for _, k := range rawKlines {
		var v StockItem
		time1, _ := timeFromUnixTimestampFloat(k[0])

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

		items_final = append(items_final, v)

	}
	if len(items_final) > 0 {
		var f = items_final[0]
		var l = items_final[len(items_final)-1]
		var candel StockItem
		candel.D = f.D
		candel.T = f.T

		candel.O = f.O
		candel.H = f.H

		candel.L = l.L
		candel.C = l.C

		for _, it := range items_final {
			candel.V = candel.V + it.V
		}
		items_final := items_final[:0]
		items_final = append(items_final, candel)

		OutToCSVFile(items_final, dir_cache_path, item.File_name, false)
	}

	return nil
}

var global_rawKlines = [][]interface{}{}

func GetAssetYear(asset_name string, start time.Time, end time.Time, timeframe ETimeFrame, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(cachePath, asset_name, timeframe.String())
	last_candel := path.Join(cachePath, cacheLastCandel, asset_name, (D1).String()+".csv")
	var items_final []StockItem

	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)
	diff := end.Sub(start) / ((24 * time.Hour) * 360)

	for i := 0; i <= int(diff); i++ {
		var tt = start.AddDate(i, 0, 0)

		var d1 TimeRange
		d1.File_name = TimeToString(tt, "yyyy") + ".csv"
		y, _, _ := tt.Date()
		d1.Begin = time.Date(y, 1, 1, 0, 0, 0, 0, tt.Location())

		d1.End = time.Date(y, 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())

		day_rang = append(day_rang, d1)
	}

	for _, item := range day_rang {
		items_final = items_final[:0]

		start_str := strconv.FormatInt(UnixMilli(item.Begin), 10)
		end_str := strconv.FormatInt(UnixMilli(item.End), 10)

		var final_out = path.Join(dir_cache_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if IsExist(final_out) && item.Begin.Year() == time.Now().Year() {
			var err = os.Remove(final_out)
			if err != nil {
				fmt.Println(err)
				return err
			}
		} else if IsExist(final_out) {
			continue
		}
		fmt.Printf(">>> %v %v %v \n", asset_name, timeframe.String(), TimeToString(item.Begin, "yyyy-mm-dd"))

		err := GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe.String()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
		if err != nil {
			return err
		}

		for _, k := range rawKlines {
			var v StockItem
			time1, _ := timeFromUnixTimestampFloat(k[0])

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

			items_final = append(items_final, v)

		}
		if len(items_final) > 0 {
			if !OutToCSVFile(items_final, dir_cache_path, item.File_name, false) {
				return errors.New("get asset daily >>> out to csv failed")
			}
		}
	}

	if !JoinCSVFiles(dir_cache_path, file_list, final_out, last_candel) {
		return errors.New("get asset daily >>> join last candel failed")
	}

	return nil
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
	cacheFolderPath := path.Join(cachePath, assetName, (timefram).String())
	var itemsFinal []StockItem
	startStr := strconv.FormatInt(UnixMilli(item.Begin), 10)
	endStr := strconv.FormatInt(UnixMilli(item.End), 10)
	err := GetJson("https://api.binance.com/api/v3/klines?symbol="+assetName+"&interval="+timefram.String()+"&startTime="+startStr+"&endTime="+endStr, &_rawKlines)

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
		file_15m := path.Join(cachePath, assetName, (M15).String(), it.File_name)
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
			filePath15 := path.Join(cachePath, assetName, (M15).String(), last_File_date_ProcessinM15)
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
		file_15m := path.Join(cachePath, assetName, (baseTimeFrame).String(), it.File_name)

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
