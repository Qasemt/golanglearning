package stockwork

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"../helper"
)

// Open_time          int64  `json:"0"`
// O                  string `json:"1"`
// H                  string `json:"2"`
// L                  string `json:"3"`
// C                  string `json:"4"`
// V                  string `json:"5"`
// Close_time         int64  `json:"6"`
// Quote_asset_volume string `json:"7"`
// Number_of_trades   int64  `json:"8"`
// Rev1               string `json:"9"`
// Rev2               string `json:"10"`
// Rev3               string `json:"11"`
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

func GetAsset(asset_name string, start time.Time, end time.Time, timeframe string, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_path string
	dir_path = "d:\\cache\\" + asset_name + "\\" + timeframe

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

		var final_out = path.Join(dir_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if item.Begin.Year() == time.Now().Year() && item.Begin.Month() == time.Now().Month() && item.Begin.Day() == time.Now().Day() {
			var err = os.Remove(final_out)
			if err != nil {
				return err
			}
		}

		if _, err := os.Stat(final_out); !os.IsNotExist(err) {
			continue
		}

		err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
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
			helper.OutToCSVFile(items_final, dir_path, item.File_name, false)
		}
	}
	helper.JoinCSVFiles(dir_path, file_list, final_out)

	return nil
}

func GetAssetYear(asset_name string, start time.Time, end time.Time, timeframe string, final_out string) error {

	rawKlines := [][]interface{}{}
	day_rang := []TimeRange{}
	file_list := []string{}
	var dir_path string
	dir_path = "d:\\cache\\" + asset_name + "\\" + timeframe

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

		var final_out = path.Join(dir_path, item.File_name)
		file_list = append(file_list, item.File_name)

		if helper.IsExist(final_out) {
			if item.Begin.Year() == time.Now().Year() {
				var err = os.Remove(final_out)
				if err != nil {
					return err
				}
			}

			continue
		}

		err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval="+timeframe+"&startTime="+start_str+"&endTime="+end_str, &rawKlines)
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
			helper.OutToCSVFile(items_final, dir_path, item.File_name, false)
		}
	}
	helper.JoinCSVFiles(dir_path, file_list, final_out)

	return nil
}
