package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"github.com/bitlum/connector/crpc"
	"github.com/go-errors/errors"
	"strings"
)

func printJSON(resp interface{}) {
	b, err := json.Marshal(resp)
	if err != nil {
		fatal(err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	out.WriteString("\n")
	out.WriteTo(os.Stdout)
}

func printRespJSON(resp proto.Message) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}

var createReceiptCommand = cli.Command{
	Name:     "createreceipt",
	Category: "Receipt",
	Usage:    "Generates new receipt.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "asset",
			Usage: "Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "Media is a type of technology which is used to transport" +
				" value of underlying asset",
		},
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which should be received on this " +
				"receipt.",
		},
		cli.StringFlag{
			Name: "description",
			Usage: "(optional) Description works only for lightning invoices." +
				" This description will be placed in the invoice itself, " +
				"which would allow user to see what he paid for later in" +
				" the wallet.",
		},
	},
	Action: createReceipt,
}

func createReceipt(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media       crpc.Media
		asset       crpc.Asset
		amount      string
		description string
	)

	switch {
	case ctx.IsSet("media"):
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	default:
		return errors.New("media argument missing")
	}

	switch {
	case ctx.IsSet("asset"):
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	default:
		return errors.Errorf("asset argument missing")
	}

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("description") {
		description = ctx.String("description")
	}

	ctxb := context.Background()
	resp, err := client.CreateReceipt(ctxb, &crpc.CreateReceiptRequest{
		Asset:       asset,
		Media:       media,
		Amount:      amount,
		Description: description,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var validateReceiptCommand = cli.Command{
	Name:     "validatereceipt",
	Category: "Receipt",
	Usage:    "Validates given receipt.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "receipt",
			Usage: "Receipt is either blockchain address or lightning network.",
		},
		cli.StringFlag{
			Name:  "asset",
			Usage: "Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "Media is a type of technology which is used to transport" +
				" value of underlying asset",
		},
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which should be received on this " +
				"receipt.",
		},
	},
	Action: validateReceipt,
}

func validateReceipt(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media   crpc.Media
		asset   crpc.Asset
		amount  string
		receipt string
	)

	switch {
	case ctx.IsSet("media"):
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	default:
		return errors.New("media argument missing")
	}

	switch {
	case ctx.IsSet("asset"):
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	default:
		return errors.Errorf("asset argument missing")
	}

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("receipt") {
		receipt = ctx.String("receipt")
	} else {
		return errors.Errorf("receipt argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.ValidateReceipt(ctxb, &crpc.ValidateReceiptRequest{
		Asset:   asset,
		Media:   media,
		Amount:  amount,
		Receipt: receipt,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var balanceCommand = cli.Command{
	Name:     "balance",
	Category: "Balance",
	Usage:    "Return asset balance.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "asset",
			Usage: "(optional) Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "(optional) Media is a type of technology which is used to" +
				" transport value of underlying asset",
		},
	},
	Action: balance,
}

func balance(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media crpc.Media
		asset crpc.Asset
	)

	switch {
	case ctx.IsSet("media"):
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	}

	switch {
	case ctx.IsSet("asset"):
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	}

	ctxb := context.Background()
	resp, err := client.Balance(ctxb, &crpc.BalanceRequest{
		Asset: asset,
		Media: media,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var estimateFeeCommand = cli.Command{
	Name:     "estimatefee",
	Category: "Fee",
	Usage:    "Estimates fee of the payment.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "asset",
			Usage: "Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "Media is a type of technology which is used to transport" +
				" value of underlying asset",
		},
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which will be sent by" +
				" service.",
		},
		cli.StringFlag{
			Name: "receipt",
			Usage: "(optional) Receipt represent either blockchains address" +
				" or lightning network invoice. If receipt is specified the " +
				"number are more accurate for lightning network media",
		},
	},
	Action: estimateFee,
}

func estimateFee(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media   crpc.Media
		asset   crpc.Asset
		amount  string
		receipt string
	)

	switch {
	case ctx.IsSet("media"):
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "bl", "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "li", "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	default:
		return errors.New("media argument missing")
	}

	switch {
	case ctx.IsSet("asset"):
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	default:
		return errors.Errorf("asset argument missing")
	}

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	}

	if ctx.IsSet("receipt") {
		receipt = ctx.String("receipt")
	}

	ctxb := context.Background()
	resp, err := client.EstimateFee(ctxb, &crpc.EstimateFeeRequest{
		Asset:   asset,
		Media:   media,
		Amount:  amount,
		Receipt: receipt,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var sendPaymentCommand = cli.Command{
	Name:     "sendpayment",
	Category: "Payment",
	Usage:    "Sends payment",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "asset",
			Usage: "Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "Media is a type of technology which is used to transport" +
				" value of underlying asset",
		},
		cli.StringFlag{
			Name: "amount",
			Usage: "(optional) Amount is the amount which will be sent by" +
				" service.",
		},
		cli.StringFlag{
			Name: "receipt",
			Usage: "Receipt is either blockchain address or lightning network" +
				" invoice which identifies the receiver of the payment.",
		},
	},
	Action: sendPayment,
}

func sendPayment(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media   crpc.Media
		asset   crpc.Asset
		amount  string
		receipt string
	)

	switch {
	case ctx.IsSet("media"):
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "bl", "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "li", "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	default:
		return errors.New("media argument missing")
	}

	switch {
	case ctx.IsSet("asset"):
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	default:
		return errors.Errorf("asset argument missing")
	}

	if ctx.IsSet("amount") {
		amount = ctx.String("amount")
	} else if media == crpc.Media_BLOCKCHAIN {
		// In case of blockchain we always should specify amount.
		// In case of lighnting we might not do that if it specified in the
		// invoice.g
		return errors.Errorf("amount argument is missing")
	}

	if ctx.IsSet("receipt") {
		receipt = ctx.String("receipt")
	} else {
		return errors.Errorf("receipt argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.SendPayment(ctxb, &crpc.SendPaymentRequest{
		Asset:   asset,
		Media:   media,
		Amount:  amount,
		Receipt: receipt,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var paymentByIDCommand = cli.Command{
	Name:     "paymentbyid",
	Category: "Payment",
	Usage:    "Return payment by the given id",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "id",
			Usage: "ID it is unique identificator of the payment. " +
				"In case of blockchain media payment id is the transaction" +
				" id, in case of lightning media it is the payment hash.",
		},
	},
	Action: paymentByID,
}

func paymentByID(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var id string

	if ctx.IsSet("id") {
		id = ctx.String("id")
	} else {
		return errors.Errorf("id argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.PaymentByID(ctxb, &crpc.PaymentByIDRequest{
		PaymentId: id,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var paymentByReceiptCommand = cli.Command{
	Name:     "paymentbyreceipt",
	Category: "Payment",
	Usage:    "Return payment by the given receipt",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "receipt",
			Usage: "Receipt is either blockchain address or lightning network" +
				" invoice which identifies the receiver of the payment.",
		},
	},
	Action: paymentByReceipt,
}

func paymentByReceipt(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var receipt string

	if ctx.IsSet("receipt") {
		receipt = ctx.String("receipt")
	} else {
		return errors.Errorf("receipt argument is missing")
	}

	ctxb := context.Background()
	resp, err := client.PaymentsByReceipt(ctxb, &crpc.PaymentsByReceiptRequest{
		Receipt: receipt,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}

var listPaymentsCommand = cli.Command{
	Name:     "listpayments",
	Category: "Payment",
	Usage:    "Return list payments by the given filter parameters",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "asset",
			Usage: "Asset is an acronym of the crypto currency",
		},
		cli.StringFlag{
			Name: "media",
			Usage: "Media is a type of technology which is used to transport" +
				" value of underlying asset",
		},
		cli.StringFlag{
			Name: "direction",
			Usage: "Direction identifies the direction of the payment, " +
				"(incoming, outgoing, internal).",
		},
		cli.StringFlag{
			Name: "status",
			Usage: "Status is the state of the payment, " +
				"(waiting, pending, completed, failed).",
		},
	},
	Action: listPayments,
}

func listPayments(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	var (
		media     crpc.Media
		asset     crpc.Asset
		status    crpc.PaymentStatus
		direction crpc.PaymentDirection
	)

	if ctx.IsSet("media") {
		stringMedia := ctx.String("media")
		switch stringMedia {
		case "bl", "blockchain":
			media = crpc.Media_BLOCKCHAIN
		case "li", "lightning":
			media = crpc.Media_LIGHTNING
		default:
			return errors.Errorf("invalid media type %v, support media type "+
				"are: 'blockchain' and 'lightning'", stringMedia)
		}
	}

	if ctx.IsSet("asset") {
		stringAsset := strings.ToLower(ctx.String("asset"))
		switch stringAsset {
		case "btc", "bitcoin":
			asset = crpc.Asset_BTC
		case "bch", "bitcoincash":
			asset = crpc.Asset_BCH
		case "ltc", "litecoin":
			asset = crpc.Asset_LTC
		case "eth", "ethereum":
			asset = crpc.Asset_ETH
		case "dash":
			asset = crpc.Asset_DASH
		default:
			return errors.Errorf("invalid asset %v, supported assets"+
				"are: 'btc', 'bch', 'dash', 'eth', 'ltc'", stringAsset)
		}
	}

	if ctx.IsSet("status") {
		stringStatus := strings.ToLower(ctx.String("status"))
		switch stringStatus {
		case strings.ToLower(crpc.PaymentStatus_WAITING.String()):
			status = crpc.PaymentStatus_WAITING

		case strings.ToLower(crpc.PaymentStatus_PENDING.String()):
			status = crpc.PaymentStatus_PENDING

		case strings.ToLower(crpc.PaymentStatus_COMPLETED.String()):
			status = crpc.PaymentStatus_COMPLETED

		case strings.ToLower(crpc.PaymentStatus_FAILED.String()):
			status = crpc.PaymentStatus_FAILED
		default:
			return errors.Errorf("invalid status %v, supported statuses"+
				"are: 'waiting', 'pending', 'completed', 'failed'",
				stringStatus)
		}
	}

	if ctx.IsSet("direction") {
		stringDirection := strings.ToLower(ctx.String("direction"))
		switch stringDirection {
		case strings.ToLower(crpc.PaymentDirection_INTERNAL.String()):
			direction = crpc.PaymentDirection_INTERNAL

		case strings.ToLower(crpc.PaymentDirection_OUTGOING.String()):
			direction = crpc.PaymentDirection_OUTGOING

		case strings.ToLower(crpc.PaymentDirection_INCOMING.String()):
			direction = crpc.PaymentDirection_INCOMING

		default:
			return errors.Errorf("invalid direction %v, supported direction"+
				"are: 'incoming', 'outgoing', 'internal'",
				stringDirection)
		}
	}

	ctxb := context.Background()
	resp, err := client.ListPayments(ctxb, &crpc.ListPaymentsRequest{
		Status:    status,
		Direction: direction,
		Asset:     asset,
		Media:     media,
	})
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}
