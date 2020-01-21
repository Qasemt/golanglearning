package avardstock

import (
	"errors"
	"fmt"
	. "github.com/qasemt/helper"
	"net/http"
	"net/url"
	"strconv"
	"sync"
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
//-===========================
type Jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
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
	err := GetJson("https://rahavard365.com/api/chart/bars?ticker=exchange."+isAssetStr+"%3A"+sq.Stock.AssetCode+"%3Areal_close"+typechart+"&resolution="+frame+"&startDateTime="+startStr+"&endDateTime="+endStr+"&firstDataRequest=true", &raws,&a.HttpLock)

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

