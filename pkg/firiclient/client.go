package firiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

type Doer func(*http.Request) (*http.Response, error)

func New(base *url.URL, httpClient Doer) *publicClient {
	return &publicClient{
		baseurl: base,
		doer:    httpClient,
	}
}

func NewAuthenticatedClient(base *url.URL, s *signer, publicClient *publicClient, doer Doer) *authClient {
	return &authClient{
		publicClient: publicClient,
		baseurl:      base,
		doer:         doer,
		signer:       s,
	}
}

type authClient struct {
	*publicClient
	signer  *signer
	baseurl *url.URL
	doer    Doer
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
func (c *publicClient) GetMarketTickerV2(ctx context.Context, marketId string) (*MarketTicker, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + marketId + "/ticker")
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
		m.MarketID = marketId
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

// GET /v2/balances
func (c *authClient) GetBalancesV2(ctx context.Context) (*Balances, error) {
	uri, err := c.baseurl.Parse("/v2/balances")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", uri.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.doSigned(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		m := Balances{}
		err = json.Unmarshal(body, &m)
		return &m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
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
	OrderType string    `json:"type"`
	Amount    float64   `json:"amount,string"`
	Price     float64   `json:"price,string"`
	Total     float64   `json:"total,string"`
	CreatedAt time.Time `json:"created_at"`
}

// GET /v2/markets/:market/history
func (c *publicClient) GetMarketTradeHistoryV2(ctx context.Context, marketId string) (*TradeHistory, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + marketId + "/history")
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
func (c *publicClient) GetOrderbookV2(ctx context.Context, marketId string) (*Orderbook, error) {
	uri, err := c.baseurl.Parse("/v2/markets/" + marketId + "/depth")
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

func (c *authClient) doSigned(r *http.Request) (*http.Response, error) {
	now := time.Now()
	sig, err := c.signer.Sign(now)
	if err != nil {
		return nil, err
	}
	r.Header.Set("miraiex-user-clientid", sig.ClientID)
	r.Header.Set("miraiex-user-signature", sig.Signature)
	uri := *r.URL
	q := uri.Query()
	q.Set("timestamp", strconv.FormatInt(sig.Timestamp.Unix(), 10))
	q.Set("validity", strconv.FormatInt(sig.ValidForMillis, 10))
	uri.RawQuery = q.Encode()
	r.URL = &uri

	return c.do(r)
}

func (c *publicClient) do(r *http.Request) (*http.Response, error) {
	log := zerolog.Ctx(r.Context())
	start := time.Now()
	uri := r.URL.String()
	log.Info().Str("method", r.Method).Str("uri", uri).Msgf("%v: %v -->", r.Method, uri)

	res, err := c.doer(r)
	elapsed := time.Since(start)
	if err != nil {
		log.Warn().
			Str("method", r.Method).
			Str("uri", uri).
			Dur("elapsed", elapsed).
			Err(err).
			Msgf("%v: %v <-- ERROR: %v", r.Method, uri, err)
		return res, err
	} else {
		log.Info().
			Str("method", r.Method).
			Str("uri", uri).
			Int("status", res.StatusCode).
			Dur("elapsed", elapsed).
			Msgf("%v: %v <-- %v in %vms", r.Method, uri, res.Status, elapsed.Milliseconds())
		return res, err
	}
}
