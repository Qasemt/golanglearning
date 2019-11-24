package stockwork

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"os"
	"path"
	"strconv"
	"strings"
	"time"


)




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



func GetAsset(assetName string, start time.Time, end time.Time, timeframe ETimeFrame) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(GetRootCache(),"data" ,"crypto", assetName, timeframe.ToString())
	last_candel := path.Join(GetRootCache(),"data" ,"crypto", cacheLastCandel, assetName, (timeframe).ToString()+".csv")
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

		fmt.Printf(">>> %v %v %v \n", assetName, timeframe.ToString(), TimeToString(item.Begin, "yyyy-mm-dd"))
		err := GetJson("https://api.binance.com/api/v3/klines?symbol="+assetName+"&interval="+timeframe.ToString()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
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
	 filePath := path.Join(GetRootCache(), "crypto", assetName,fmt.Sprintf("%v_%v.csv", strings.ToLower(assetName), strings.ToLower(timeframe.ToString2())) )

	if !JoinCSVFiles(dir_cache_path, file_list,filePath , last_candel) {
		return errors.New("get asset >>> join last candel failed")
	}

	return nil
}
func GetAssetCreateLastCandel(assetName string, end time.Time, timeframe ETimeFrame) error {

	rawKlines := [][]interface{}{}
	var dir_cache_path string
	dir_cache_path = path.Join(GetRootCache(),"data" ,"crypto", cacheLastCandel, assetName)

	var items_final []StockItem

	var item TimeRange
	item.File_name = timeframe.ToString() + ".csv"
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

func GetAssetYear(asset_name string, start time.Time, end time.Time, timeframe ETimeFrame) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(GetRootCache(),"data" ,"crypto", asset_name, timeframe.ToString())
	last_candel := path.Join(GetRootCache(),"data" ,"crypto", cacheLastCandel, asset_name, (D1).ToString()+".csv")
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
		fmt.Printf(">>> %v %v %v \n", asset_name, timeframe.ToString(), TimeToString(item.Begin, "yyyy-mm-dd"))

		err := GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe.ToString()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
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
	filePath := path.Join(GetRootCache(), "crypto", asset_name,fmt.Sprintf("%v_%v.csv", strings.ToLower(asset_name), strings.ToLower(timeframe.ToString2())) )

	if !JoinCSVFiles(dir_cache_path, file_list, filePath, last_candel) {
		return errors.New("get asset daily >>> join last candel failed")
	}

	return nil
}


func Make(asset string) error {
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO

	var begin time.Time
	var end time.Time
	end = (time.Now())
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO minute
	//f := time.Minute * -30
	//begin = (end.Add(f))

	now1 := time.Now()
	e1 := GetAssetCreateLastCandel(asset, now1, M15)
	if e1 != nil {
		return e1
	}
	e2 := GetAssetCreateLastCandel(asset, now1, H1)
	if e2 != nil {
		return e2
	}

	e3 := GetAssetCreateLastCandel(asset, now1, H2)
	if e3 != nil {
		return e3
	}
	e4 := GetAssetCreateLastCandel(asset, now1, H4)
	if e4 != nil {
		return e4
	}
	e5 := GetAssetCreateLastCandel(asset, now1, D1)
	if e5 != nil {
		return e5
	}

	//:::::::::::::::::::::::::::::::::::::::: CRYPTO HOUR
	begin = (end.AddDate(0, 0, -30))
	e6 := GetAsset(asset, begin, end, M15)
	if e6 != nil {
		return e6
	}

	begin = (end.AddDate(0, 0, -30))
	e7 := GetAsset(asset, begin, end, H1)
	if e7 != nil {
		return e7
	}

	begin = (end.AddDate(0, 0, -30))
	e8 := GetAsset(asset, begin, end, H2)
	if e8 != nil {
		return e8
	}
	begin = (end.AddDate(0, 0, -100))
	e9 := GetAsset(asset, begin, end, H4)
	if e9 != nil {
		return e9
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO DAILY
	begin = (end.AddDate(-2, 0, 0))
	e10 := GetAssetYear(asset, begin, end, D1 )
	if e10 != nil {
		return e10
	}
	//:::::::::::::::::::::::::::::::::::::::: CRYPTO 1 min
	return nil
}
