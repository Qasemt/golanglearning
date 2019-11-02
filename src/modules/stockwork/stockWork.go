package stockwork

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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

//::::::::::::::::::::::::::::::::::::::::::::::::::::::: DEFINE
var path_watch_list string = "./tehran_watch_list.json"
var path_src_dir string = ""
var path_dst_dir string = ""
var log = logrus.New()

func toInt(s string) int64 {

	f, e := strconv.ParseInt(s, 10, 64)

	if e != nil {
		return -1
	}

	return f
}
func CastInt(s string) string {

	f, e := strconv.ParseFloat(s, 64)

	if e != nil {
		return "0"
	}

	return strconv.FormatInt(int64(f), 10)
}
func CastFloat(f string) string {
	var res string
	f = strings.Replace(f, " ", "", -1)
	if s, err := strconv.ParseFloat(f, 64); err == nil {
		//res = fmt.Sprintf("%0.0000f", s)
		res = strconv.FormatFloat(s, 'f', 4, 64)

	}
	return res
}
func toFloat(f string) float64 {
	var res float64

	f = strings.Replace(f, " ", "", -1)
	res, err := strconv.ParseFloat(f, 64)
	if err != nil {
		res = 0
	}
	return res
}
func toINT64(v string) int64 {
	var res int64

	v = strings.Replace(v, " ", "", -1)
	res, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		res = 0
	}
	return res
}

//::::::::::::::::::::::::::::::::::::::::::::::::::::::::: FUNCATIONs
type watchListItems struct {
	Qlist []watchListItem `json:"q"`
}

type watchListItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type stockRecord struct {
	DATE    string
	TIME    string
	OPEN    string
	HIGH    string
	LOW     string
	CLOSE   string
	VOLUME  string
	OPENINT string
}

func (c stockRecord) toString(is_cast bool) []string {
	s := make([]string, 7)
	s[0] = c.DATE
	s[1] = c.TIME
	if is_cast {
		s[2] = CastInt(c.OPEN)
		s[3] = CastInt(c.HIGH)
		s[4] = CastInt(c.LOW)
		s[5] = CastInt(c.CLOSE)
		s[5] = CastInt(c.VOLUME)
		s[6] = CastInt(c.OPENINT)
	} else {
		s[2] = c.OPEN
		s[3] = c.HIGH
		s[4] = c.LOW
		s[5] = c.CLOSE
		s[5] = c.VOLUME
		s[6] = c.OPENINT
	}
	return s
}

func readJsonWatchList(wlist *watchListItems) bool {
	jsonFile, err := os.Open(path_watch_list)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &wlist)

	return true
}

func readCsvFile(filePath string) ([][]string, []stockRecord, error) {
	f, _ := os.Open(filePath)
	var s [][]string
	var list []stockRecord
	r := csv.NewReader(f)

	for {

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
			//panic(err)
		}
		s = append(s, record)
		f := stockRecord{
			DATE:    record[0],
			TIME:    "000000",
			OPEN:    record[2],
			HIGH:    record[3],
			LOW:     record[4],
			CLOSE:   record[5],
			VOLUME:  record[6],
			OPENINT: record[7],
		}
		list = append(list, f)
	}
	return s, list, nil
}

func csvExport(data []stockRecord, out string) error {
	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.UseCRLF = true
	for _, value := range data {
		var final []string

		if strings.Contains(value.DATE, "<time>") {
			final = value.toString(false)
		} else {
			final = value.toString(true)
		}
		if err := writer.Write(final); err != nil {
			return err // let's return errors if necessary, rather than having a one-size-fits-all error handler
		}
	}
	return nil
}

func logInit() {
	const LogFilePath = "logs/misc.log"

	// log.SetFormatter(&helper.Formatter{
	// 	HideKeys:        false,
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// 	NoColors:        true,
	// })

	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) //default
	log.Formatter.(*logrus.TextFormatter).TimestampFormat = "2006-01-02 15:04:05"
	log.Formatter.(*logrus.TextFormatter).DisableColors = true     // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = false // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	lumberjackLogrotate := &lumberjack.Logger{
		Filename:   LogFilePath,
		MaxSize:    1,  // Max megabytes before log is rotated
		MaxBackups: 2,  // Max number of old log files to keep
		MaxAge:     60, // Max number of days to retain log files
		Compress:   true,
	}

	logMultiWriter := io.MultiWriter(os.Stdout, lumberjackLogrotate)
	log.SetOutput(logMultiWriter)

	log.WithFields(logrus.Fields{
		"Runtime Version": runtime.Version(),
		"Number of CPUs":  runtime.NumCPU(),
		"Arch":            runtime.GOARCH,
	}).Info("Application Initializing")
}

func RUNStock(src_dir string, dst_dir string, is_adj bool) {

	path_src_dir = src_dir
	path_dst_dir = dst_dir
	//:::::::::::::::::::::::::::: Setup  LOGGER ::::::::::::::::::::::::::::
	logInit()
	if is_adj {
		log.Infof("->>>------ ADJUST---> ")
	}
	//::::::::::::::::::::::::::::  requirement ::::::::::::::::::::::::::::
	if _, err := os.Stat(path_watch_list); os.IsNotExist(err) {

		log.Errorf("file not exist :[ %v ]\n", path_watch_list)
		return
	}

	if _, err := os.Stat(path_src_dir); os.IsNotExist(err) {

		log.Errorf("dir not exist :[ %v ]\n", path_src_dir)
		return
	}

	var stockList watchListItems

	readJsonWatchList(&stockList)

	var f_dst string
	var f_src string
	for _, g := range stockList.Qlist {

		f_dst = ""
		f_src = ""

		if is_adj && strings.HasPrefix(g.Id, "IRX") {
			f_src = path.Join(path_src_dir, g.Id+".csv")
			f_dst = path.Join(path_dst_dir, g.Name+"_adj"+".csv")
		} else {
			if is_adj {
				f_src = path.Join(path_src_dir, g.Id+"-i.csv")
				f_dst = path.Join(path_dst_dir, g.Name+"_adj"+".csv")
			} else {
				f_src = path.Join(path_src_dir, g.Id+".csv")
				f_dst = path.Join(path_dst_dir, g.Name+".csv")
			}
		}

		var _, list, e = readCsvFile(f_src)

		if e != nil {
			if os.IsNotExist(e) {
				log.Printf("File Does Not Exist:[%v]\n", f_src)
			} else {
				log.Errorf("failed %v:[%v]\n", g.Name, f_src)
			}
		}

		if _, err := os.Stat(path_dst_dir); os.IsNotExist(err) {
			os.MkdirAll(path_dst_dir, os.ModePerm)
		}

		if _, err := os.Stat(f_dst); !os.IsNotExist(err) {

			var err = os.Remove(f_dst)
			if err != nil {
				log.Printf("remove failed :[%v][%v]\n", err.Error(), f_dst)
			}
		}

		var err = csvExport(list, f_dst)

		if err != nil {
			log.Printf("export failed [%v]\n", f_dst)
		} else {

			log.WithFields(logrus.Fields{
				"adj": is_adj,
			}).Info("success >>> " + f_dst)
		}

	}
	log.Info("finished >>>")

}

func ConvertStoockTODT7(src_file_csv string, dst_file_csv string) {

	var final_out = dst_file_csv
	if _, err := os.Stat(src_file_csv); os.IsNotExist(err) {

		log.Errorf("dir not exist :[ %v ]\n", src_file_csv)
		return
	}

	f, _ := os.Open(src_file_csv)

	if dst_file_csv == "" {
		var fname = strings.Split(filepath.Base(src_file_csv), ".")[0] + "_out.csv"

		var src_dir_name = filepath.Dir(src_file_csv)

		final_out = path.Join(src_dir_name, fname)

		fmt.Println(fname, src_dir_name, final_out)
	}
	var s [][]string
	var list []stockRecord
	r := csv.NewReader(f)
	var i = 0
	for {

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
			//panic(err)
		}
		s = append(s, record)
		var f stockRecord
		if i == 0 {
			f = stockRecord{
				DATE:    record[0],
				TIME:    "<TIME>",
				OPEN:    record[1],
				HIGH:    record[2],
				LOW:     record[3],
				CLOSE:   record[4],
				VOLUME:  "VOLUME",
				OPENINT: "OPEN",
			}
		} else {
			f = stockRecord{
				DATE:    record[0],
				TIME:    "000000",
				OPEN:    record[1],
				HIGH:    record[2],
				LOW:     record[3],
				CLOSE:   record[4],
				VOLUME:  "000000",
				OPENINT: "0",
			}
			// if record[4]!=""{
			//  f.VOLUME = record[3]
			// }
		}
		list = append(list, f)
		i++
	}

	//:::::::::::::::::::::::::::::::::::::

	file, err := os.Create(final_out)
	if err != nil {
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	//for i := len(s) - 1; i >= 0; i-- {
	for i := 0; i < len(s); i++ {
		//for _, value := range list {
		value := list[i]
		var final []string

		if i == 0 {
			final = value.toString(false)
		} else {
			final = make([]string, 8)
			// t, e := time.Parse(
			// 	time.RFC3339,
			// 	value.DATE)
			// fmt.Println(t, e)

			//var s []string
			//s = strings.Split(value.DATE, "/")
			//	value.DATE = strconv.FormatInt(int64((toInt(s[2])+2000)), 10) + "" + s[0] + "" + s[1]

			value.DATE = strings.Replace(value.DATE, "-", "", -1)

			final[0] = value.DATE
			final[1] = value.TIME
			final[2] = CastFloat(value.OPEN)
			final[3] = CastFloat(value.HIGH)
			final[4] = CastFloat(value.LOW)
			final[5] = CastFloat(value.CLOSE)
			final[6] = CastFloat(value.VOLUME)
			final[7] = value.OPENINT
		}
		writer.UseCRLF = true
		if err := writer.Write(final); err != nil {
			return // let's return errors if necessary, rather than having a one-size-fits-all error handler
		}
	}
	return

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

func outToCSVFile(items []StockItem, dst_file_csv string) bool {

	var final_out = dst_file_csv

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

func getJson(url_path string, target_object_json interface{}) error {

	fmt.Println(url_path)
	fixedURL, err := url.Parse("http://103.21.150.184:3128")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return err
	}

	transport := &http.Transport{Proxy: http.ProxyURL(fixedURL)}

	var myClient = &http.Client{Timeout: 15 * time.Second, Transport: transport}
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
func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

type StockItemBinance struct {
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

	Open_time          int64
	O                  string
	H                  string
	L                  string
	C                  string
	V                  string
	Close_time         int64
	Quote_asset_volume string
	Number_of_trades   int64
	Rev1               string
	Rev2               string
	Rev3               string
}
type ResultBinanceStock struct {
	Result []StockItemBinance
}

func GetJsonBTC(url string) {

	var items_from_binance [][]string
	var items_final []StockItem
	var start_num time.Time
	var end time.Time

	end = (time.Now())
	start_num = (time.Now().AddDate(0, 0, -1))

	start_str := strconv.FormatInt(unixMilli(start_num), 10)
	end_str := strconv.FormatInt(unixMilli(end), 10)
	//threeDays := time.Hour * 24 * 3
	//	diff := now.Add(threeDays)

	getJson("https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1h&startTime="+start_str+"&endTime="+end_str, &items_from_binance)
	for i := 0; i < len(items_from_binance); i++ {
		var v StockItem
		v.D = items_from_binance[i][0]
		v.T = items_from_binance[i][0]

		v.O = toFloat(items_from_binance[i][1])
		v.H = toFloat(items_from_binance[i][2])
		v.L = toFloat(items_from_binance[i][3])
		v.C = toFloat(items_from_binance[i][4])
		v.V = toFloat(items_from_binance[i][5])
		items_final = append(items_final, v)
	}

	outToCSVFile(items_final, "d:\\tt.csv")
	fmt.Println(items_from_binance)

}
