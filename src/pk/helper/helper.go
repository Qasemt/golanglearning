package helper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	socks "github.com/samuel/go-socks/socks"
)

var url_proxy string
var _apikey string
var _secret string

var is_Socks bool
var mRootCachePath string
var _verboe bool

//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: set
func SetProxy(v string, is_socks bool) error {

	/*_, err := url.Parse(v)
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return err
	}*/
	url_proxy = v
	is_Socks = is_socks
	return nil
}
func SetSecret(v string) {
	_secret = v
}
func SetAPIKey(v string) {
	_apikey = v
}
func SetRootCache(p string) {
	mRootCachePath = p
}
func SetVerbose(b bool){
	_verboe =b;
}
//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: get
func GetRootCache() string {
	if mRootCachePath == "" {
		mRootCachePath = "./d/"
	}
	return mRootCachePath
}
func GetProxy() string {
	return url_proxy
}
func GetSecret() string {
	return _secret
}
func GetAPIKey() string {
	return _apikey
}
func GetVerbose() bool{
	return _verboe;
}
//_______________________________________________________________________ create File watch list and info dt
/*txtinfo.inf*/
func CreateconfigTxtinfo(outPath string) error {

	s := path.Join(outPath, "txtinfo.inf")
	if IsExist(s) {
		return nil
	}
	f, err := os.Create(s)
	if err != nil {
		return err
	}

	defer f.Close()
	_, e := f.WriteString("delimit  = ,\r\nskip     =   1\r\ndt       =  1\r\nti       =  2\r\nop       =  3\r\nhi       =  4\r\nlo       =  5\r\ncl       =  6\r\nvol      =  7\r\noi       =  0\r\ndtformat = CCYYMMDD\r\ntiformat = HHMMSS\r\next      = csv\r\ncf       = 4\r\n")
	if e != nil {
		return e
	}
	f.Sync()

	return nil
}
func CreateWatchList(outPath string) error {
	f :=false
	s := path.Join(outPath, "watchList.json")
	if IsExist(s) {
		return nil
	}
	data := WatchListItem{

		Tehran: []WatchStock{
			{
				NameEn:"vaghadir",
				AssetCode:"66",
				IsIndex: false,
				TimeFrame:[]string{},
				IsAdj:&f,
			},
			{
				NameEn:"Senosa",
				AssetCode:"425",
				IsIndex: false,
				TimeFrame:[]string{},
				IsAdj:&f,
			},
			{
				NameEn:"sefars",
				AssetCode:"400",
				IsIndex: false,
				TimeFrame:[]string{},
				IsAdj:&f,
			},
			{
				NameEn:"vanovin",
				AssetCode:"393",
				IsIndex: false,
				TimeFrame:[]string{},
				IsAdj:&f,
			},

		},

		Crypto: []WatchStock{
			{
				NameEn:"BTCUSDT",
				AssetCode:"BTCUSDT",
				IsIndex: false,
				TimeFrame:[]string{},
			},


		},
	}

	file, _ := json.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile(s, file, 0644)

	return nil
}

func UnixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
func UnixTimeToTime(millis int64) QTime {
	//return time.Unix(0, millis*int64(time.Millisecond))
	tm := QTime{time.Unix(millis, 0)}
	return tm
}
func IsExist(p string) bool {
	res := false
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		res = true
	}
	return res
}
func CompareDate(d1 time.Time, d2 time.Time) bool {
	y, m, d := d1.Date()
	y22, m22, d22 := d2.Date()
	if y == y22 && m == m22 && d == d22 {
		return true
	}
	return false
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
	} else if format == "yyyy-mm-dd" {
		formatted = fmt.Sprintf("%4d-%02d-%02d",
			t.Year(), t.Month(), t.Day())
	} else {
		formatted = fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	}
	return formatted
}

func UnixTimeStrToFormatDT(t time.Time, is_date bool,tf ETimeFrame) string {
	var formatted string
	if is_date {

		formatted = fmt.Sprintf("%4d%02d%02d",
			t.Year(), t.Month(), t.Day())
	} else {
		if tf ==D1 {
			formatted = fmt.Sprintf("%02d%02d%02d",
				t.Hour(), t.Minute(), 0)
		}else {
			formatted = fmt.Sprintf("%02d%02d%02d",
				t.Hour(), t.Minute(),t.Second())
		}
	}

	return formatted
}

func UnixTimeStrToTimeFormat(millis string, is_date bool) string {
	var gg = ToINT64(millis)
	t := UnixTimeToTime(gg)
	var formatted string
	if is_date {

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
func ToINT32(v string) int32 {

	res, err :=strconv.Atoi(v)
	if err != nil {
		res = 0
	}
	return int32(res)
}

func GetJson(url_path string, target_object_json interface{},mux *sync.Mutex) error {
	mux.Lock()
	if GetVerbose() {
		fmt.Println("GetJson -> ", url_path)
	}
	defer mux.Unlock()
	//https://github.com/binance-exchange/go-binance/blob/1af034307da53bf592566c5c8a90856ddb5b34a4/util.go#L49
	//fmt.Println(url_path)
	var _timeout time.Duration= 60

	var myClient *http.Client
	if GetProxy() != "" {

		if myClient == nil {
			if !is_Socks {
				fixedURL, err := url.Parse(GetProxy())
				if err != nil {
					fmt.Println("Malformed URL: ", err.Error())
					return err
				}
				transport := &http.Transport{Proxy: http.ProxyURL(fixedURL)}
				myClient = &http.Client{Timeout: _timeout * time.Second, Transport: transport}
			} else {
				proxy := &socks.Proxy{GetProxy(), "", ""}
				tr := &http.Transport{
					Dial: proxy.Dial,
				}
				myClient = &http.Client{Timeout: _timeout * time.Second, Transport: tr}
			}
		}

	} else {
		var myTransport http.RoundTripper = &http.Transport{
		//	Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: time.Second * _timeout,
		}
		//myClient = &http.Client{Timeout: 120 * time.Second}
		myClient = &http.Client{Transport: myTransport}
	}
	req, err := http.NewRequest("GET", url_path, nil)
	if err != nil {
		fmt.Println("Req -> : ", err.Error())
		return err
	}
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Cookie", "_ga=GA1.2.2095670304.1547216406; _hjid=5dd39770-33ce-46d9-9979-a95820cfdb24; .rahavard365auth=23917073533FDBF44324DBDD56B7AC020AFD4221A05ECBC4C7BD09AA8B25016CB9A430AB550F236EB9B55ECB88A3629EA1FD02C85306F806C4B88E272E1B1E69DC62AF5EC38B78FE1FE20CA4A89687024E0CBF2BFBE31065C66B6A7E886EEBBA96A24EFC37830CF60F488A0CB69C3FE82A8ED3A664129F430C628B3A4B9542023517EC045ACBADAED4EBE87AA151DC564096087D563C704FF746134BF44CDBE995401844724BCB99F2643335A497245A4214452DBE81EC6765DF53B11E79B60E67894357D2151D2B8BB9740D827D86CF448ED640E33732BF2895C28A0602945A38C71F104538C07FD011420D64CA361D; _gid=GA1.2.275768049.1580750965")
	resp, err := myClient.Do(req)
	if err != nil {
		return  err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return  err
	}
	if resp.StatusCode != http.StatusOK {
	//	return  fmt.Errorf("status %d: %v", resp.StatusCode, string(response))
		return  fmt.Errorf("HTTP Error Status %v ",resp.StatusCode)
	}

	json.Unmarshal(response, &target_object_json)
	return  err
}
func GetJsonBin(url_path string, target_object_json interface{},mux *sync.Mutex) error {
	//fmt.Println(url_path)
	if GetVerbose() {
		fmt.Println("GetJson -> ", url_path)
	}
	c := &ClientHelper{
		window: 5000,
		apikey: GetAPIKey(),
		secret: GetSecret(),
		client: http.DefaultClient,
	}
	res, err := c.do(http.MethodGet, url_path, nil, false, false,mux)

	if err != nil {
		return err
	}

	err1 := json.Unmarshal(res, &target_object_json)
	if err1 != nil {
		return err1
	}

	return nil
}
func OutToCSVFile(items []StockItem, dir_path string, dst_file_csv string, is_add_header bool) bool {

	if _, err := os.Stat(dir_path); os.IsNotExist(err) {
		os.MkdirAll(dir_path, os.ModePerm)
	}
	if dst_file_csv == "" {
		fmt.Println("OutToCSVFile", "dest file name is empty :(")
		return false
	}

	var final_out = path.Join(dir_path, dst_file_csv)
	CreateconfigTxtinfo(dir_path)
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

func AppendToCSVFile(items []StockItem, dst_file_csv string, is_add_header bool) bool {

	if dst_file_csv == "" {
		fmt.Println("AppendToCSVFile", "dest file name is empty :(")
		return false
	}
	if !IsExist(dst_file_csv) {
		fmt.Println("file not found ", dst_file_csv)
		return false
	}

	//:::::::::::::::::::::::::::::::::::::

	fmain, err := os.OpenFile(dst_file_csv, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	writer := csv.NewWriter(fmain)
	writer.UseCRLF = true
	defer fmain.Close()
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
func JoinCSVFiles(dir_path string, dst_file_csv_list []string, out_final_file string, last_value_from_1m string) bool {

	if _, err := os.Stat(dir_path); os.IsNotExist(err) {
		return false
	}
	a := filepath.Dir(out_final_file)
	if _, err := os.Stat(a); os.IsNotExist(err) {
		merr := os.MkdirAll(a, os.ModePerm)
		if merr != nil {
			return false
		}
	}

	file, err := os.Create(out_final_file)
	if err != nil {
		fmt.Println(fmt.Sprintf("joinCsvFiles() failed -> ", out_final_file))
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
	header1[7] = "<OPENINT>"

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
	//---------- last value from timeframe 1minute
	f1, _ := os.Open(last_value_from_1m)

	// Create a new reader.
	r1 := csv.NewReader(f1)
	defer f1.Close()
	for {
		record, err := r1.Read()
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
	return true

}
func JoinTwoCSVFiles(mainFilePath string, secondFilePath string) bool {

	if !IsExist(mainFilePath) {
		fmt.Println("file not found ", mainFilePath)
		return false
	}
	if !IsExist(secondFilePath) {
		fmt.Println("file not found ", secondFilePath)
		return false
	}

	fread, _ := os.Open(secondFilePath)
	//-----------------
	fmain, err := os.OpenFile(mainFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// Create a new reader.
	r1 := csv.NewReader(fread)
	defer fread.Close()

	writer := csv.NewWriter(fmain)
	writer.UseCRLF = true
	defer fmain.Close()
	defer writer.Flush()
	for {
		record, err := r1.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		if err != nil {
			return false
		}
		if err := writer.Write(record); err != nil {
			fmt.Println("JoinTwoCSVFiles :failed")
			return false
		}
	}
	return true
}
func GetJsonToArry(mainFilePath string) (bool, [][]string) {

	if !IsExist(mainFilePath) {
		fmt.Println("file not found ", mainFilePath)
		return false, nil
	}
	var list [][]string

	fread, err := os.Open(mainFilePath)

	if err != nil {
		return false, nil
	}
	// Create a new reader.
	r1 := csv.NewReader(fread)
	defer fread.Close()

	for {
		record, err := r1.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, nil
		}
		list = append(list, record)

	}
	return true, list
}

func Round(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

func RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

func RoundDown(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Floor(digit)
	newVal = round / pow
	return
}