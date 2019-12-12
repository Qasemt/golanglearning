package avardstock

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"strconv"
)
type TehranLoader struct {
	StockProvider
}


func NewTehran(store_mode EFolderStoreMode) *TehranLoader {
	t := TehranLoader{StockProvider{}}

	t.StockProvider.IStockProvider = &t
	t.Provider = Avard
	t.FolderStoreMode = store_mode
	return &t
}
func (a TehranLoader) downloadAsset(sq StockQuery, item TimeRange) ([]StockFromWebService, error) {
	var _rawKines = []StockFromWebService{}
	startStr := strconv.FormatInt(item.Begin.Unix(), 10)
	endStr := strconv.FormatInt(item.End.Unix(), 10)
	var frame string
	if sq.TimeFrame == D1 {
		frame = "D"
	} else if sq.TimeFrame == M15 {
		frame = "15"
	} else if sq.TimeFrame == H1 {
		frame = "60"
	} else if sq.TimeFrame == H2 {
		frame = "120"
	} else if sq.TimeFrame == H4 {
		frame = "240"
	}
	var typechart string = ""
	var isAssetStr string = "asset" //asset / index
	if sq.TypeChart == Adj {
		typechart = "%3Atype1"
	}
	if sq.Stock.IsIndex == true {
		isAssetStr = "index"
	}
	//var raws []interface{}
	var raws []stocktemp
	var itemsFinal []StockFromWebService
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange."+isAssetStr+"%3A"+sq.Stock.AssetCode+"%3Areal_close"+typechart+"&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &raws)

	if err != nil {
		return nil, err
	}

	if _rawKines == nil {
		return nil, errors.New(fmt.Sprintf("downloadAsset failed ... %v\n", err))
	}

	for _, k := range raws {
		var v StockFromWebService
		v.Time = int64(k.Time)
		v.O = k.O
		v.H = k.H
		v.L = k.L
		v.C = k.C
		v.V = k.V
		itemsFinal = append(itemsFinal, v)

	}

	return itemsFinal, nil
}