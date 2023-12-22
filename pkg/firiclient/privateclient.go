package firiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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

type ActiveOrder struct {
	Id        int64     `json:"id"`
	Market    string    `json:"market"`
	Type      OrderType `json:"type"`
	Price     float64   `json:"price,string"`
	Amount    float64   `json:"amount,string"`
	Remaining float64   `json:"remaining,string"`
	Matched   float64   `json:"matched,string"`
	Cancelled float64   `json:"cancelled,string"`
	CreatedAt time.Time `json:"created_at"`
}

type ActiveOrders []ActiveOrder

// GET /v2/orders
func (c *authClient) GetActiveOrders(ctx context.Context) (ActiveOrders, error) {
	uri, err := c.baseurl.Parse("/v2/orders")
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
		m := ActiveOrders{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// GET /v2/orders/history
func (c *authClient) GetAllFilledAndClosedOrders(ctx context.Context) (ActiveOrders, error) {
	uri, err := c.baseurl.Parse("/v2/orders/history")
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
		m := ActiveOrders{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

type HistoricTrade struct {
	Id             string    `json:"id"`
	Market         string    `json:"market"`
	Price          float64   `json:"price,string"`
	PriceCurrency  string    `json:"price_currency"`
	Amount         float64   `json:"amount,string"`
	AmountCurrency string    `json:"amount_currency"`
	Cost           float64   `json:"cost,string"`
	CostCurrency   string    `json:"cost_currency"`
	Side           string    `json:"side"`
	IsMaker        bool      `json:"isMaker"`
	Date           time.Time `json:"date"`
}

type HistoricTrades []HistoricTrade

// GET /v2/history/trades
func (c *authClient) GetAllTrades(ctx context.Context) (HistoricTrades, error) {
	uri, err := c.baseurl.Parse("/v2/history/trades")
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
		m := HistoricTrades{}
		err = json.Unmarshal(body, &m)
		return m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// GET /v2/orders/:marketId
func (c *authClient) GetActiveOrdersInMarket(ctx context.Context, marketId MarketID) (*ActiveOrders, error) {
	uri, err := c.baseurl.Parse("/v2/orders/" + string(marketId))
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
		m := ActiveOrders{}
		err = json.Unmarshal(body, &m)
		return &m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

// DELETE /v2/orders
func (c *authClient) DeleteAllOrders(ctx context.Context) (*ActiveOrders, error) {
	uri, err := c.baseurl.Parse("/v2/orders")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", uri.String(), nil)
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
		m := ActiveOrders{}
		err = json.Unmarshal(body, &m)
		return &m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

type CreateOrderResponse struct {
	Id int64 `json:"id"`
}
type CreateOrderRequest struct {
	Market string    `json:"market"`
	Type   OrderType `json:"type"`
	Price  float64   `json:"price,string"`
	Amount float64   `json:"amount,string"`
}

// POST /v2/orders
func (c *authClient) PostOrder(ctx context.Context, r *CreateOrderRequest) (*CreateOrderResponse, error) {
	uri, err := c.baseurl.Parse("/v2/orders")
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.doSigned(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		m := CreateOrderResponse{}
		err = json.Unmarshal(body, &m)
		return &m, err
	} else {
		return nil, fmt.Errorf("%v: %v: status=%v body=%v", req.Method, uri.String(), resp.StatusCode, string(body))
	}
}

type CreateWithdrawalRequest struct {
	Amount  string `json:"amount"`
	Address string `json:"address"`
}
type CreateWithdrawalResponse struct{}

// POST /v2/withdraw/:coin
func (c *authClient) PostWithdrawal(ctx context.Context, coinId string, r *CreateWithdrawalRequest) (*CreateWithdrawalResponse, error) {
	uri, err := c.baseurl.Parse("/v2/withdraw/" + coinId)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.doSigned(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		m := CreateWithdrawalResponse{}
		err = json.Unmarshal(body, &m)
		return &m, err
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
	r.Header.Set("miraiex-access-key", c.signer.apiKey)
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
