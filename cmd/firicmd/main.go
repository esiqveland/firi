package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/esiqveland/firi/pkg/firiclient"
)

var firiApiUrl = getEnv("FIRI_API_URL", "https://api.firi.com")
var apiKeyEnv = mustGetEnv("FIRI_API_KEY")
var clientIdEnv = mustGetEnv("FIRI_CLIENT_ID")
var secretKeyEnv = mustGetSecret("FIRI_SECRET_KEY")

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

	baseUrl, err := url.Parse(firiApiUrl)
	if err != nil {
		return err
	}

	httpclient := &http.Client{Timeout: time.Second * 5}
	signer := firiclient.NewSigner(
		clientIdEnv,
		apiKeyEnv,
		secretKeyEnv,
	)
	publicClient := firiclient.New(
		baseUrl,
		httpclient.Do,
	)
	c := firiclient.NewAuthenticatedClient(
		baseUrl,
		signer,
		publicClient,
		httpclient.Do,
	)

	// marketsV1, err := c.GetMarketsV1(root)
	// if err != nil {
	//	return err
	// }
	// log.Printf("MarketsV1=%+v", marketsV1)
	//
	// marketsV2, err := c.GetMarketsV2(root)
	// if err != nil {
	//	return err
	// }
	// log.Printf("MarketsV2=%+v", marketsV2)
	//
	// orderbook, err := c.GetOrderbookV2(root, "BTCNOK")
	// if err != nil {
	//	return err
	// }
	// log.Printf("OrderbookV2=%+v", orderbook)
	//
	// allTickersV2, err := c.GetMarketTickersV2(root)
	// if err != nil {
	//	return err
	// }
	// log.Printf("allTickersV2=%+v", allTickersV2)
	//
	// tickerBTCNOK, err := c.GetMarketTickerV2(root, "BTCNOK")
	// if err != nil {
	//	return err
	// }
	// log.Printf("tickerBTCNOK=%+v", tickerBTCNOK)
	//
	// historyBTCNOK, err := c.GetMarketTradeHistoryV2(root, "BTCNOK")
	// if err != nil {
	//	return err
	// }
	// log.Printf("historyBTCNOK=%+v", historyBTCNOK)

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

	// tradesHistory, err := c.GetAllTrades(root)
	// if err != nil {
	// 	return err
	// }
	// log.Printf("GetAllTrades=%+v", tradesHistory)

	history, err := c.GetDepositsHistory(root)
	if err != nil {
		return err
	}
	log.Printf("GetDepositsHistory=%+v", ToJsonString(history))

	pendingTravelRule, err := c.GetPendingTravelRule(root)
	if err != nil {
		return err
	}
	log.Printf("GetPendingTravelRule=%+v", ToJsonString(pendingTravelRule))
	if len(pendingTravelRule.Deposits) > 0 {
		// take 1 and report it with vasp:
		deposit := pendingTravelRule.Deposits[0]
		log.Printf("Deposit=%+v", deposit)
		depositId := deposit["id"].(string)
		log.Printf("DepositId=%+v", depositId)
		vasp, err := c.UpdateTravelRuleVasp(root, firiclient.UpdateTravelRuleVaspRequest{
			DepositId: depositId,
			VaspDid:   robovaspAccept.DID,
			VaspName:  robovaspAccept.Name,
		})
		if err != nil {
			return err
		}
		log.Printf("UpdateTravelRuleVasp=%+v", ToJsonString(vasp))
	}

	pendingTravelRuleOriginator, err := c.GetPendingTravelRuleOriginator(root)
	if err != nil {
		return err
	}
	log.Printf("GetPendingTravelRuleOriginator=%+v", ToJsonString(pendingTravelRuleOriginator))
	if len(pendingTravelRuleOriginator.Deposits) > 0 {
		deposit := pendingTravelRuleOriginator.Deposits[0]
		log.Printf("GetPendingTravelRuleOriginator: Deposit=%+v", deposit)
		depositId := deposit["id"].(string)
		log.Printf("GetPendingTravelRuleOriginator: DepositId=%+v", depositId)
		response1, err := c.UpdateTravelRuleOriginator(root, firiclient.UpdateTravelRuleOriginatorRequest{
			DepositId: depositId,
			//Kind:      firiclient.OriginatorKindMyself,
			//Kind: "",
		})
		if err != nil {
			return err
		}
		log.Printf("UpdateTravelRuleOriginator=%+v", ToJsonString(response1))
	}

	return nil
}

func ToJsonString(object any) string {
	marshal, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return string(marshal)
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
	// data, err := base64.StdEncoding.DecodeString(val)
	// if err != nil {
	//	log.Fatal().Msgf("Error decoding secret key in ENV: %v: %v", key, err)
	// }
	// return data
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Msgf("Must set ENV: %v", key)
	}
	return val
}

func getEnv(envVal string, defaultValue string) string {
	val := os.Getenv(envVal)
	if val == "" {
		return defaultValue
	}
	return val
}

var robovaspAccept = newVasp("did:ethr:0x5cb7821acae54a5a94939fd35db91f4b11d8cd85", "RoboVASP Accept")

type Vasp struct {
	DID  string
	Name string
}

func newVasp(did string, name string) Vasp {
	return Vasp{
		DID:  did,
		Name: name,
	}
}
