package lnd

import (
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"github.com/bitlum/connector/connectors"
	"io/ioutil"
	"strconv"
	"google.golang.org/grpc/credentials"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
	"gopkg.in/macaroon.v2"
	"github.com/lightningnetwork/lnd/macaroons"
	"net"
)

var satoshiPerBitcoin = decimal.New(btcutil.SatoshiPerBitcoin, 0)

func btcToSatoshi(amount string) (int64, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, errors.Errorf("unable to parse amount(%v): %v",
			amount, err)
	}

	a, _ := amt.Float64()
	btcAmount, err := btcutil.NewAmount(a)
	if err != nil {
		return 0, errors.Errorf("unable to parse amount(%v): %v", a, err)
	}

	return int64(btcAmount), nil
}

func sat2DecAmount(amount btcutil.Amount) decimal.Decimal {
	amt := decimal.NewFromBigInt(big.NewInt(int64(amount)), 0)
	return amt.Div(satoshiPerBitcoin)
}

func generatePaymentID(invoiceStr string,
	direction connectors.PaymentDirection) string {
	return connectors.GeneratePaymentID(invoiceStr, string(direction))
}

// getClient return lightning network grpc client.
func (c *Connector) getClient(macaroonPath string) (lnrpc.LightningClient,
	*grpc.ClientConn, error) {

	creds, err := credentials.NewClientTLSFromFile(c.cfg.TlsCertPath, "")
	if err != nil {
		return nil, nil, errors.Errorf("unable to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if macaroonPath != "" {
		macaroonBytes, err := ioutil.ReadFile(macaroonPath)
		if err != nil {
			return nil, nil, errors.Errorf("Unable to read macaroon file: %v", err)
		}

		mac := &macaroon.Macaroon{}
		if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
			return nil, nil, errors.Errorf("Unable to unmarshal macaroon: %v", err)
		}

		opts = append(opts,
			grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)))
	}

	target := net.JoinHostPort(c.cfg.Host, strconv.Itoa(c.cfg.Port))
	log.Infof("lightning client connection to lnd: %v", target)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, nil, errors.Errorf("unable to to dial grpc: %v", err)
	}

	return lnrpc.NewLightningClient(conn), conn, nil
}
