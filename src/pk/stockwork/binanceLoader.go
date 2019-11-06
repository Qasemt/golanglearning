package stockwork

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"./../helper"
)

type ETimeFrame int

const (
	M1  ETimeFrame = 1
	M15 ETimeFrame = 2
	H1  ETimeFrame = 3
	H2  ETimeFrame = 4
	H4  ETimeFrame = 5
	D1  ETimeFrame = 6
)

var cache_path = "d:/cache/"
var cache_last_candel = "lastcandel"

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
	println(str)
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

type TimeRange struct {
	File_name string
	Begin     time.Time
	End       time.Time
}

func GetAsset(asset_name string, start time.Time, end time.Time, timeframe ETimeFrame, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(cache_path, asset_name, timeframe.String())
	last_candel := path.Join(cache_path, cache_last_candel, asset_name, (timeframe).String()+".csv")
	var items_final []helper.StockItem

	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)
	diff := end.Sub(start) / (24 * time.Hour)

	for i := 1; i <= int(diff); i++ {
		var tt = start.AddDate(0, 0, i)

		var d1 TimeRange
		d1.File_name = helper.TimeToString(tt, "yyyymmdd") + ".csv"
		y, m, d := tt.Date()

		d1.Begin = time.Date(y, m, d, 0, 0, 0, 0, tt.Location())
		d1.End = time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())

		day_rang = append(day_rang, d1)

	}
	for _, item := range day_rang {
		items_final = items_final[:0]

		start_str := strconv.FormatInt(helper.UnixMilli(item.Begin), 10)
		end_str := strconv.FormatInt(helper.UnixMilli(item.End), 10)

		var final_out = path.Join(dir_cache_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if helper.IsExist(final_out) && item.Begin.Year() == time.Now().Year() && item.Begin.Month() == time.Now().Month() && item.Begin.Day() == time.Now().Day() {
			var err = os.Remove(final_out)
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		if helper.IsExist(final_out) {
			continue
		}

		fmt.Println(timeframe, item)
		err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe.String()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
		if err != nil {

			fmt.Println(err)
			panic(err)
		}

		for _, k := range rawKlines {
			var v helper.StockItem
			time1, _ := timeFromUnixTimestampFloat(k[0])

			v.D = helper.UnixTimeStrToFormatDT(time1, true)
			v.T = helper.UnixTimeStrToFormatDT(time1, false)

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
			helper.OutToCSVFile(items_final, dir_cache_path, item.File_name, false)
		}
	}
	helper.JoinCSVFiles(dir_cache_path, file_list, final_out, last_candel)

	return nil
}
func GetAssetCreateLastCandel(asset_name string, end time.Time, timeframe ETimeFrame) error {

	rawKlines := [][]interface{}{}
	var dir_cache_path string
	dir_cache_path = path.Join(cache_path, cache_last_candel, asset_name)

	var items_final []helper.StockItem

	var item TimeRange
	item.File_name = timeframe.String() + ".csv"
	var diff time.Duration
	if timeframe == H1 {
		diff = time.Hour * 1
	} else if timeframe == H2 {
		diff = time.Hour * 2
	} else if timeframe == H4 {
		diff = time.Hour * 4
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
	fmt.Println(item)
	start_str := strconv.FormatInt(helper.UnixMilli(item.Begin), 10)
	end_str := strconv.FormatInt(helper.UnixMilli(item.End), 10)

	err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval=1m&startTime="+start_str+"&endTime="+end_str, &rawKlines)
	if err != nil {

		fmt.Println(err)
		return (err)
	}

	for _, k := range rawKlines {
		var v helper.StockItem
		time1, _ := timeFromUnixTimestampFloat(k[0])

		v.D = helper.UnixTimeStrToFormatDT(time1, true)
		v.T = helper.UnixTimeStrToFormatDT(time1, false)

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
		var candel helper.StockItem
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

		helper.OutToCSVFile(items_final, dir_cache_path, item.File_name, false)
	}

	return nil
}

var global_rawKlines = [][]interface{}{}

func GetAssetCreateLastCandel2(asset_name string, end time.Time, timeframe ETimeFrame) error {

	var dir_cache_path string
	dir_cache_path = path.Join(cache_path, cache_last_candel, asset_name)

	var items_final []helper.StockItem

	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
	if len(global_rawKlines) == 0 {
		var begin_time time.Time
		begin_time = end.Add(-(time.Hour * 36))
		start_str := strconv.FormatInt(helper.UnixMilli(begin_time), 10)
		end_str := strconv.FormatInt(helper.UnixMilli(end), 10)
		err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval=3m&startTime="+start_str+"&endTime="+end_str, &global_rawKlines)
		if err != nil {

			fmt.Println(err)
			return (err)
		}
	}
	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
	var item TimeRange
	item.File_name = timeframe.String() + ".csv"
	var diff time.Duration
	if timeframe == H1 {
		diff = time.Hour * 1
	} else if timeframe == H2 {
		diff = time.Hour * 2
	} else if timeframe == H4 {
		diff = time.Hour * 4
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
	fmt.Println(item)

	for _, k := range global_rawKlines {
		var v helper.StockItem
		time1, _ := timeFromUnixTimestampFloat(k[0])
		if time1.UnixNano() > item.Begin.UnixNano() {
			//	fmt.Println(">>>", timeframe, item.Begin, time1)
			v.D = helper.UnixTimeStrToFormatDT(time1, true)
			v.T = helper.UnixTimeStrToFormatDT(time1, false)

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

	}
	if len(items_final) > 0 {
		var f = items_final[0]
		var l = items_final[len(items_final)-1]
		var candel helper.StockItem
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

		helper.OutToCSVFile(items_final, dir_cache_path, item.File_name, false)
	}

	return nil
}

func GetAssetYear(asset_name string, start time.Time, end time.Time, timeframe ETimeFrame, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_cache_path string
	dir_cache_path = path.Join(cache_path, asset_name, timeframe.String())
	last_candel := path.Join(cache_path, cache_last_candel, asset_name, (D1).String()+".csv")
	var items_final []helper.StockItem

	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)
	diff := end.Sub(start) / ((24 * time.Hour) * 360)

	for i := 0; i <= int(diff); i++ {
		var tt = start.AddDate(i, 0, 0)

		var d1 TimeRange
		d1.File_name = helper.TimeToString(tt, "yyyy") + ".csv"
		y, _, _ := tt.Date()
		d1.Begin = time.Date(y, 1, 1, 0, 0, 0, 0, tt.Location())

		d1.End = time.Date(y, 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), tt.Location())

		day_rang = append(day_rang, d1)
	}

	for _, item := range day_rang {
		items_final = items_final[:0]

		start_str := strconv.FormatInt(helper.UnixMilli(item.Begin), 10)
		end_str := strconv.FormatInt(helper.UnixMilli(item.End), 10)

		var final_out = path.Join(dir_cache_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if helper.IsExist(final_out) && item.Begin.Year() == time.Now().Year() {
			var err = os.Remove(final_out)
			if err != nil {
				fmt.Println(err)
				return err
			}
		} else if helper.IsExist(final_out) {
			continue
		}
		fmt.Printf(">>> %v %v %v \n", asset_name, timeframe.String(), helper.TimeToString(item.Begin, "yyyy-mm-dd"))

		err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe.String()+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
		if err != nil {

			fmt.Println(err)
			panic(err)
		}

		for _, k := range rawKlines {
			var v helper.StockItem
			time1, _ := timeFromUnixTimestampFloat(k[0])

			v.D = helper.UnixTimeStrToFormatDT(time1, true)
			v.T = helper.UnixTimeStrToFormatDT(time1, false)

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
			helper.OutToCSVFile(items_final, dir_cache_path, item.File_name, false)
		}
	}

	helper.JoinCSVFiles(dir_cache_path, file_list, final_out, last_candel)

	return nil
}
