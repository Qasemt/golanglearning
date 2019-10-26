package testFunctions

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

var path_list string = "D:/workspace/stock/stock_data_cleaner/tehran_watch_list.json"
var path_src_dir string = "D:/workspace/stock/tseclient/normal/"
var path_dst_dir string = "D:/workspace/stock/tseclient/tmp/"

func csvExport(data [][]string, out string) error {
	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		if err := writer.Write(value); err != nil {
			return err // let's return errors if necessary, rather than having a one-size-fits-all error handler
		}
	}
	return nil
}

func readCsvFile(filePath string) ([][]string, error) {
	// Load a csv file.
	f, _ := os.Open(filePath)
	var s [][]string
	// Create a new reader.
	r := csv.NewReader(f)

	for {

		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
			//panic(err)
		}

		s = append(s, record)
		//fmt.Println(record)
		//fmt.Println(len(record))
		// for value := range record {
		// 	fmt.Printf("  %v\n", record[value])
	}

	return s, nil
}

func testCSV() bool {

	var f, e = readCsvFile("D:/workspace/stock/data/crypto/Bitcoin.csv")

	if e != nil {

		if os.IsNotExist(e) {
			fmt.Print("File Does Not Exist: ")
			return false
		}

	}

	csvExport(f, "d:/result1.csv")

	return true
}

func RunTestCSV() {
	fmt.Println(path_list)
}
