package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/esiqveland/firi/pkg/firiclient"
)

var secretKeyEnv = mustGetSecret("SECRET_KEY")
var clientIdEnv = mustGetEnv("CLIENT_ID")
var apiKeyEnv = mustGetEnv("API_KEY")

func main() {
	err := runMain()
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func runMain() error {
	logout := zerolog.ConsoleWriter{Out: os.Stderr}
	logout2 := zerolog.ConsoleWriter{Out: os.Stderr}
	logger := zerolog.New(logout2).With().Timestamp().Logger()
	log.Logger = log.Output(logout)

	root := logger.WithContext(context.Background())

	httpclient := &http.Client{Timeout: time.Second * 5}
	signer := firiclient.NewSigner(
		clientIdEnv,
		apiKeyEnv,
		secretKeyEnv,
	)
	publicClient := firiclient.New(
		mustParseUrl("https://api.miraiex.com"),
		httpclient.Do,
	)
	c := firiclient.NewAuthenticatedClient(
		mustParseUrl("https://api.miraiex.com"),
		signer,
		publicClient,
		httpclient.Do,
	)

	//marketsV1, err := c.GetMarketsV1(root)
	//if err != nil {
	//	return err
	//}
	//log.Printf("MarketsV1=%+v", marketsV1)
	//
	//marketsV2, err := c.GetMarketsV2(root)
	//if err != nil {
	//	return err
	//}
	//log.Printf("MarketsV2=%+v", marketsV2)
	//
	//orderbook, err := c.GetOrderbookV2(root, "BTCNOK")
	//if err != nil {
	//	return err
	//}
	//log.Printf("OrderbookV2=%+v", orderbook)
	//
	//allTickersV2, err := c.GetMarketTickersV2(root)
	//if err != nil {
	//	return err
	//}
	//log.Printf("allTickersV2=%+v", allTickersV2)
	//
	//tickerBTCNOK, err := c.GetMarketTickerV2(root, "BTCNOK")
	//if err != nil {
	//	return err
	//}
	//log.Printf("tickerBTCNOK=%+v", tickerBTCNOK)
	//
	//historyBTCNOK, err := c.GetMarketTradeHistoryV2(root, "BTCNOK")
	//if err != nil {
	//	return err
	//}
	//log.Printf("historyBTCNOK=%+v", historyBTCNOK)

	bal, err := c.GetBalancesV2(root)
	if err != nil {
		return err
	}
	log.Printf("Balances=%+v", bal)

	act, err := c.GetActiveOrders(root)
	if err != nil {
		return err
	}
	log.Printf("GetActiveOrders=%+v", act)

	filled, err := c.GetAllFilledAndClosedOrders(root)
	if err != nil {
		return err
	}
	log.Printf("GetAllFilledAndClosedOrders=%+v", filled)

	tradesHistory, err := c.GetAllTrades(root)
	if err != nil {
		return err
	}
	log.Printf("GetAllTrades=%+v", tradesHistory)

	return nil
}

func mustParseUrl(uri string) *url.URL {
	u, err := url.Parse(uri)
	if err != nil {
		log.Fatal().Msgf("error parsing url=%v: %v", uri, err)
	}
	return u
}

func mustGetSecret(key string) []byte {
	val := mustGetEnv(key)
	return []byte(val)
	//data, err := base64.StdEncoding.DecodeString(val)
	//if err != nil {
	//	log.Fatal().Msgf("Error decoding secret key in ENV: %v: %v", key, err)
	//}
	//return data
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Msgf("Must set ENV: %v", key)
	}
	return val
}
