package helper
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/go-querystring/query"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	url_bin       = "https://www.binance.com"
	wsAddress = "wss://stream.binance.com:9443/ws/"
)

// ClientHelper represents the actual HTTP ClientHelper, that is being used to interact with binance API server
type ClientHelper struct {
	apikey string
	secret string
	client *http.Client
	window int
}

// do invokes the given API command with the given data
// sign indicates whether the api call should be done with signed payload
// stream indicates if the request is stream related
func (c *ClientHelper) do(method, endpoint string, data interface{}, sign bool, stream bool) (response []byte, err error) {
	// Convert the given data to urlencoded format
	values, err := query.Values(data)
	if err != nil {
		return nil, err
	}

	payload := values.Encode()
	// Signed requests require the additional timestamp, window size and signature of the payload
	// Remark: This is done only to routes with actual data
	if sign {
		payload = fmt.Sprintf("%s&timestamp=%v&recvWindow=%d", payload, time.Now().UnixNano()/(1000*1000), c.window)
		mac := hmac.New(sha256.New, []byte(c.secret))
		_, err = mac.Write([]byte(payload))
		if err != nil {
			return nil, err
		}
		payload = fmt.Sprintf("%s&signature=%s", payload, hex.EncodeToString(mac.Sum(nil)))
	}

	// Construct the http request
	// Remark: GET requests payload is as a query parameters
	// POST requests payload is given as a body
	var req *http.Request
	if method == http.MethodGet {
		req, err = http.NewRequest(method, fmt.Sprintf("%s/%s", url_bin, endpoint), nil)
	} else {
		req, err = http.NewRequest(method, fmt.Sprintf("%s/%s", url_bin, endpoint), strings.NewReader(payload))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if sign || stream {
		req.Header.Add("X-MBX-APIKEY", c.apikey)
	}


	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %v", resp.StatusCode, string(response))
	}
	return response, err
}