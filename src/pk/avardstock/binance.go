package avardstock
import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"strconv"
)
type BinanceLoader struct {
	StockProvider
}

func NewBinance(store_mode EFolderStoreMode) *BinanceLoader {
	t := BinanceLoader{StockProvider{}}
	t.StockProvider.IStockProvider = &t
	t.Provider = Binance
	t.FolderStoreMode = store_mode
	return &t
}

func (a BinanceLoader) downloadAsset(sq StockQuery, item TimeRange) ([]StockFromWebService, error) {
	var _rawKines = []StockFromWebService{}
	startStr := strconv.FormatInt(UnixMilli(item.Begin), 10)
	endStr := strconv.FormatInt(UnixMilli(item.End), 10)
	rawKlines := [][]interface{}{}
	var itemsFinal []StockFromWebService
	err := GetJsonBin("api/v3/klines?symbol="+sq.Stock.AssetCode+"&interval="+sq.TimeFrame.ToString()+"&startTime="+startStr+"&endTime="+endStr, &rawKlines)

	if err != nil {
		return nil, err
	}

	if _rawKines == nil {
		return nil, errors.New(fmt.Sprintf("downloadAsset failed ... %v\n", err))
	}

	for _, k := range rawKlines {
		var v StockFromWebService
		ts, _ := k[0].(float64)
		v.Time = int64(ts)
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

	return itemsFinal, nil
}
