package estimator

import (
	"fmt"
	"net/http"

	"sync"
	"time"

	"io/ioutil"

	"encoding/json"

	core "github.com/bitlum/viabtc_rpc_client"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// USDEstimator estimates the given asset in usd dollar.
type USDEstimator interface {
	// Estimate is used to estimate asset.
	Estimate(asset string, amount string) (string, error)
}

var acronimToName = map[string]string{
	string(core.AssetBTC):  "bitcoin",
	string(core.AssetBCH):  "bitcoin-cash",
	string(core.AssetETH):  "ethereum",
	string(core.AssetLTC):  "litecoin",
	string(core.AssetDASH): "dash",
}

// Coinmarket is an asset estimator which is based on coinmarketcap data.
// It periodically fetches the last price from website and makes the
// estimation based on this prices.
type Coinmarket struct {
	quit chan struct{}
	wg   sync.WaitGroup

	info     map[string]decimal.Decimal
	commands chan *estimateCommand
}

// A compile time check to ensure Coinmarket implements the USDEstimator
// interface.
var _ USDEstimator = (*Coinmarket)(nil)

// NewCoinmarketcapEstimator returns new instance of coinmarketcap estimator.
func NewCoinmarketcapEstimator() *Coinmarket {
	return &Coinmarket{
		info:     make(map[string]decimal.Decimal),
		quit:     make(chan struct{}),
		commands: make(chan *estimateCommand),
	}
}

// Start start estimation service.
func (e *Coinmarket) Start() error {
	e.wg.Add(1)
	go func() {
		defer func() {
			e.wg.Done()
			log.Info("Quit updating goroutine")
		}()

		// Make an initial price fetching to avoid
		for asset, name := range acronimToName {
			select {
			case <-e.quit:
			default:
			}

			if err := e.update(asset, name); err != nil {
				log.Errorf("unable to update price %v: %v", name, err)
			}
		}

		for {
			select {
			case <-time.After(time.Second * 30):
				for asset, name := range acronimToName {
					select {
					case <-e.quit:
					default:
					}

					if err := e.update(asset, name); err != nil {
						log.Errorf("unable to update price %v: %v", name, err)
					}
				}
			case cmd := <-e.commands:
				amount, err := decimal.NewFromString(cmd.amount)
				if err != nil {
					cmd.err <- errors.Errorf("unable parse float amount: %v", err)
					continue
				}

				if price, ok := e.info[cmd.asset]; ok {
					cmd.res <- price.Mul(amount).String()
					continue
				} else {
					cmd.err <- errors.Errorf("estimation of asset %v is not supported",
						cmd.asset)
					continue
				}
			case <-e.quit:
				return
			}
		}
	}()

	return nil
}

func (e *Coinmarket) Stop() {
	close(e.quit)
	e.wg.Wait()
}

func (e *Coinmarket) update(asset, name string) error {
	url := fmt.Sprintf("https://api.coinmarketcap."+
		"com/v1/ticker/%v/?convert=USD", name)

	resp, err := http.Get(url)
	if err != nil {
		return errors.Errorf("unable to make request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("unable to read response body: %v", err)
	}

	type Resp struct {
		PriceUsd string `json:"price_usd"`
	}

	var data []Resp
	if err := json.Unmarshal(body, &data); err != nil {
		return errors.Errorf("unable to decode response body: %v", err)
	}

	if len(data) == 0 {
		return errors.Errorf("response not contains data")
	}

	if data[0].PriceUsd == "" {
		return errors.New("unable to find price field")
	}

	price, err := decimal.NewFromString(data[0].PriceUsd)
	if err != nil {
		return errors.Errorf("unable parse float price: %v", err)
	}

	e.info[asset] = price

	log.Infof("Update %v price: %v", name, price.String())
	return nil
}

type estimateCommand struct {
	asset  string
	amount string
	res    chan string
	err    chan error
}

// Estimate estimates the given amount with usd price from coinmarketcap.
// Esimation function will be working only after initial price fetching, so it
// could be safely used right after start.
func (e *Coinmarket) Estimate(asset string, amount string) (string, error) {
	res := make(chan string)
	errChan := make(chan error)

	select {
	case e.commands <- &estimateCommand{
		asset:  asset,
		amount: amount,
		res:    res,
		err:    errChan,
	}:
	case <-e.quit:
		return "", errors.Errorf("service is shutdown")
	}

	select {
	case total := <-res:
		return total, nil
	case err := <-errChan:
		return "", err
	case <-e.quit:
		return "", errors.Errorf("service is shutdown")
	}
}
