package firiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type OrderType string

const (
	Bid OrderType = "bid"
	Ask OrderType = "ask"
)

var orderTypes = map[string]OrderType{
	"bid": Bid,
	"ask": Ask,
}

func (o OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(o))
}

func (o *OrderType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	val, ok := orderTypes[s]
	if !ok {
		return errors.New("json: invalid orderType value=" + s)
	}
	*o = val
	return nil
}

type MarketID string

const (
	BTCNOK MarketID = "BTCNOK"
	ETHNOK MarketID = "ETHNOK"
	DAINOK MarketID = "DAINOK"
	ADANOK MarketID = "ADANOK"
	LTCNOK MarketID = "LTCNOK"
)

type Doer func(*http.Request) (*http.Response, error)

func New(base *url.URL, httpClient Doer) *publicClient {
	return &publicClient{
		baseurl: base,
		doer:    httpClient,
	}
}

type publicClient struct {
	baseurl *url.URL
	doer    Doer
}

type Markets []Market
type Market struct {
	ID     string  `json:"id"`
	Last   float64 `json:"last,string"`
	High   float64 `json:"high,string"`
	Change float64 `json:"change,string"`
	Low    float64 `json:"low,string"`
	Volume float64 `json:"volume,string"`
}

// GET /v1/markets
func (c *publicClient) GetMarketsV1(ctx context.Context) (Markets, error) {
	uri, err := c.baseurl.Parse("/v1/markets")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := Markets{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// GET /v2/markets
func (c *publicClient) GetMarketsV2(ctx context.Context) (Markets, error) {
	uri, err := c.baseurl.Parse("/v2/markets")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := Markets{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

type MarketTickers []MarketTicker
type MarketTicker struct {
	MarketID string  `json:"market"`
	Bid      float64 `json:"bid,string"`
	Ask      float64 `json:"ask,string"`
	Spread   float64 `json:"spread,string"`
}

// GET /v2/markets/tickers
func (c *publicClient) GetMarketTickersV2(ctx context.Context) (MarketTickers, error) {
	uri, err := c.baseurl.Parse("/v2/markets/tickers")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := MarketTickers{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// GET /v2/markets/:market/ticker
func (c *publicClient) GetMarketTickerV2(ctx context.Context, marketId MarketID) (*MarketTicker, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + string(marketId) + "/ticker")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := MarketTicker{}
		err = json.Unmarshal(body, &m)
		m.MarketID = string(marketId)
		return &m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

type Balances []Balance
type Balance struct {
	MarketID string  `json:"id"`
	Last     float64 `json:"last,string"`
	High     float64 `json:"high,string"`
	Low      float64 `json:"low,string"`
	Change   float64 `json:"change,string"`
	Volume   float64 `json:"volume,string"`
}

type Bids []ordersJsonList
type Asks []ordersJsonList

// ordersJsonList: a order is a list of ["price", "quantity"], both float64 as string
type ordersJsonList []interface{}

func (s ordersJsonList) ToOrder() (Order, bool) {
	o := Order{}
	if len(s) != 2 {
		return o, false
	}
	priceStr := s[0].(string)
	quantityStr := s[1].(string)
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return o, false
	}
	quantity, err := strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		return o, false
	}
	o.Price = price
	o.Quantity = quantity
	return o, true
}

type Order struct {
	Price    float64
	Quantity float64
}
type Orderbook struct {
	Bids []Order
	Asks []Order
}
type orderbookJson struct {
	Bids Bids `json:"bids"`
	Asks Asks `json:"asks"`
}

type TradeHistory []HistoricOrder
type HistoricOrder struct {
	OrderType OrderType `json:"type"`
	Amount    float64   `json:"amount,string"`
	Price     float64   `json:"price,string"`
	Total     float64   `json:"total,string"`
	CreatedAt time.Time `json:"created_at"`
}

// GET /v2/markets/:market/history
func (c *publicClient) GetMarketTradeHistoryV2(ctx context.Context, marketId MarketID) (*TradeHistory, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + string(marketId) + "/history")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := &TradeHistory{}
		err = json.Unmarshal(body, m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// GET /v2/markets/:market/depth
func (c *publicClient) GetOrderbookV2(ctx context.Context, marketId MarketID) (*Orderbook, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + string(marketId) + "/depth")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := orderbookJson{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			return nil, err
		}
		bids := make([]Order, len(m.Bids))
		asks := make([]Order, len(m.Asks))

		for i := range m.Bids {
			o, ok := m.Bids[i].ToOrder()
			if !ok {
				return nil, fmt.Errorf("error parsing val=%+v", m.Bids[i])
			}
			bids[i] = o
		}
		for i := range m.Asks {
			o, ok := m.Asks[i].ToOrder()
			if !ok {
				return nil, fmt.Errorf("error parsing val=%+v", m.Asks[i])
			}
			asks[i] = o
		}
		o := &Orderbook{
			Bids: bids,
			Asks: asks,
		}
		return o, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

func (c *publicClient) do(r *http.Request) (*http.Response, error) {
	log := zerolog.Ctx(r.Context())
	start := time.Now()
	uri := r.URL.String()
	if r.Header.Get("x-request-id") == "" {
		xId := xid.New().String()
		r.Header.Set("x-request-id", xId)
	}
	xId := r.Header.Get("x-request-id")
	log.Info().Str("method", r.Method).Str("uri", uri).Str("x-request-id", xId).Msgf("%v: %v -->", r.Method, uri)

	res, err := c.doer(r)
	elapsed := time.Since(start)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", r.Method).
			Str("uri", uri).
			Dur("elapsed", elapsed).
			Str("x-request-id", xId).
			Msgf("%v: %v <-- ERROR: %v", r.Method, uri, err)
		return res, err
	} else {
		if res.StatusCode > 300 {
			log.Warn().
				Str("method", r.Method).
				Str("uri", uri).
				Int("status", res.StatusCode).
				Dur("elapsed", elapsed).
				Str("x-request-id", xId).
				Msgf("%v: %v <-- %v in %vms", r.Method, uri, res.Status, elapsed.Milliseconds())
			return res, err
		} else {
			log.Info().
				Str("method", r.Method).
				Str("uri", uri).
				Int("status", res.StatusCode).
				Dur("elapsed", elapsed).
				Str("x-request-id", xId).
				Msgf("%v: %v <-- %v in %vms", r.Method, uri, res.Status, elapsed.Milliseconds())
			return res, err
		}
	}
}
