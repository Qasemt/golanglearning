package stockwork

import (
	"fmt"
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
func timeFromUnixTimestampFloat(raw interface{}) (time.Time, error) {
	ts, ok := raw.(float64)
	if !ok {
		return time.Time{}, nil
	}
	return time.Unix(0, int64(ts)*int64(time.Millisecond)), nil
}

func GetAsset(url string, asset_name string, start time.Time, end time.Time) {

	rawKlines := [][]interface{}{}
	//var items_from_binance [][]string

	var items_final []helper.StockItem

	start_str := strconv.FormatInt(helper.UnixMilli(start), 10)
	end_str := strconv.FormatInt(helper.UnixMilli(end), 10)

	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)

	err := helper.GetJson("https://api.binance.com/api/v3/klines?symbol="+asset_name+"&interval=1h&startTime="+start_str+"&endTime="+end_str, &rawKlines)
	if err != nil {

		fmt.Println(err)
		panic(err)
	}
	//klines := []*Kline{}
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
	// for i := 0; i < len(rawKlines); i++ {
	// 	t := rawKlines[i]
	// 	var v helper.StockItem

	// 	v.D = helper.UnixTimeStrToTimeFormat(t[0], true)
	// 	v.T = helper.UnixTimeStrToTimeFormat(t[0], false)

	// 	v.O = helper.ToFloat(t[1])
	// 	v.H = helper.ToFloat(t[2])
	// 	v.L = helper.ToFloat(t[3])
	// 	v.C = helper.ToFloat(t[4])
	// 	v.V = helper.ToFloat(t[5])
	// 	items_final = append(items_final, v)
	// }

	//	helper.OutToCSVFile(items_final, "d:\\tt.csv")
	//fmt.Println(items_from_binance)
	fmt.Println(items_final)

}
