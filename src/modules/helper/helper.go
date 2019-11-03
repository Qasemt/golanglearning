package helper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var url_proxy string

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func SetProxy(v string) error {

	_, err := url.Parse(v)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return err
	}
	url_proxy = v
	return nil
}

func GetProxy() string {
	return url_proxy
}
func UnixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func UnixTimeToTime(millis int64) time.Time {
	//return time.Unix(0, millis*int64(time.Millisecond))
	tm := time.Unix(millis, 0)
	return tm
}
func TimeToString(t time.Time, format string) string {
	var formatted string
	if format == "yyyymmdd" {
		formatted = fmt.Sprintf("%4d%02d%02d",
			t.Year(), t.Month(), t.Day())
	} else if format == "yyyymmdd" {
		formatted = fmt.Sprintf("%4d_%02d_%02d",
			t.Year(), t.Month(), t.Day())
	} else if format == "yyyy" {
		formatted = fmt.Sprintf("%4d",
			t.Year())
	} else {
		formatted = fmt.Sprintf("%4d-%02d-%02d",
			t.Year(), t.Month(), t.Day())
	}
	return formatted
}

func UnixTimeStrToFormatDT(t time.Time, is_date bool) string {
	var formatted string
	if is_date {

		formatted = fmt.Sprintf("%4d%02d%02d",
			t.Year(), t.Month(), t.Day())
	} else {
		formatted = fmt.Sprintf("%02d%02d%02d",
			t.Hour(), t.Minute(), t.Second())
	}

	return formatted
}

func UnixTimeStrToTimeFormat(millis string, is_date bool) string {
	var gg = ToINT64(millis)
	t := UnixTimeToTime(gg)
	var formatted string
	if is_date {
		fmt.Println(t, t.Year())
		formatted = fmt.Sprintf("%4d%2d%2d",
			t.Year(), t.Month(), t.Day())
	} else {
		formatted = fmt.Sprintf("%2d%02d%02d",
			t.Hour(), t.Minute(), t.Second())
	}

	return formatted
}

func ToFloat(f string) float64 {
	var res float64

	f = strings.Replace(f, " ", "", -1)
	res, err := strconv.ParseFloat(f, 64)
	if err != nil {
		res = 0
	}
	return res
}
func ToINT64(v string) int64 {
	var res int64

	v = strings.Replace(v, " ", "", -1)
	res, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		res = 0
	}
	return res
}

func GetJson(url_path string, target_object_json interface{}) error {
	//https://github.com/binance-exchange/go-binance/blob/1af034307da53bf592566c5c8a90856ddb5b34a4/util.go#L49
	fmt.Println(url_path)
	var myClient *http.Client
	if GetProxy() != "" {
		fixedURL, err := url.Parse(GetProxy())
		if err != nil {
			fmt.Println("Malformed URL: ", err.Error())
			return err
		}
		transport := &http.Transport{Proxy: http.ProxyURL(fixedURL)}

		myClient = &http.Client{Timeout: 30 * time.Second, Transport: transport}
	} else {
		myClient = &http.Client{Timeout: 30 * time.Second}
	}
	r, err := myClient.Get(url_path)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(body, &target_object_json)
	//fmt.Printf("body len : %v\n %v\n", len(body), string(body))
	return err
	//return json.NewDecoder(r.Body).Decode(target)
}

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

func OutToCSVFile(items []StockItem, dir_path string, dst_file_csv string, is_add_header bool) bool {

	if _, err := os.Stat(dir_path); os.IsNotExist(err) {
		os.MkdirAll(dir_path, os.ModePerm)
	}

	var final_out = path.Join(dir_path, dst_file_csv)

	//var s [][]string

	//:::::::::::::::::::::::::::::::::::::

	file, err := os.Create(final_out)
	if err != nil {
		return false
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.UseCRLF = true
	defer writer.Flush()

	if is_add_header {
		header1 := make([]string, 8)
		header1[0] = "<DATE>"
		header1[1] = "<TIME>"
		header1[2] = "<OPEN>"
		header1[3] = "<HIGH>"
		header1[4] = "<LOW>"
		header1[5] = "<CLOSE>"
		header1[6] = "<VOLUME>"
		header1[7] = "<OPEN>"

		if err := writer.Write(header1); err != nil {
			return false
		}
	}

	for i := 0; i < len(items); i++ {

		value := items[i]

		final := make([]string, 8)

		if value.D != "" {
			final[0] = value.D
		} else {
			final[0] = "000000"
		}

		if value.T != "" {
			final[1] = value.T
		} else {
			final[1] = "000000"
		}

		final[2] = strconv.FormatFloat(value.O, 'f', 4, 64)
		final[3] = strconv.FormatFloat(value.H, 'f', 4, 64)
		final[4] = strconv.FormatFloat(value.L, 'f', 4, 64)
		final[5] = strconv.FormatFloat(value.C, 'f', 4, 64)
		final[6] = strconv.FormatFloat(value.V, 'f', 4, 64)
		final[7] = strconv.FormatFloat(value.BV, 'f', 4, 64)

		if err := writer.Write(final); err != nil {
			return false
		}
	}
	return true

}
func JoinCSVFiles(dir_path string, dst_file_csv_list []string, out_final_file string) bool {

	if _, err := os.Stat(dir_path); os.IsNotExist(err) {
		return false
	}
	file, err := os.Create(out_final_file)
	if err != nil {
		return false
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.UseCRLF = true
	defer writer.Flush()

	header1 := make([]string, 8)
	header1[0] = "<DATE>"
	header1[1] = "<TIME>"
	header1[2] = "<OPEN>"
	header1[3] = "<HIGH>"
	header1[4] = "<LOW>"
	header1[5] = "<CLOSE>"
	header1[6] = "<VOLUME>"
	header1[7] = "<OPEN>"

	if err := writer.Write(header1); err != nil {
		return false
	}

	for i := 0; i < len(dst_file_csv_list); i++ {

		var final_out = path.Join(dir_path, dst_file_csv_list[i])

		if _, err := os.Stat(final_out); os.IsNotExist(err) {
			continue
		}
		f, _ := os.Open(final_out)

		// Create a new reader.
		r := csv.NewReader(f)

		for {
			record, err := r.Read()
			// Stop at EOF.
			if err == io.EOF {
				break
			}

			if err != nil {
				return false
			}
			if err := writer.Write(record); err != nil {
				return false
			}

		}

	}
	return true

}
