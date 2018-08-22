package main

import (
	"fmt"
	"os"
	"github.com/bitlum/connector/crpc"
	"github.com/btcsuite/btcutil"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

const (
	defaultRPCPort     = "9002"
	defaultRPCHostPort = "localhost:" + defaultRPCPort
)

var (
	// Commit stores the current commit hash of this build. This should be
	// set using -ldflags during compilation.
	Commit string

	defaultLndDir = btcutil.AppDataDir("payserver", false)
)

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[lncli] %v\n", err)
	os.Exit(1)
}

func getClient(ctx *cli.Context) (crpc.PayServerClient, func()) {
	conn := getClientConn(ctx, false)

	cleanUp := func() {
		conn.Close()
	}

	return crpc.NewPayServerClient(conn), cleanUp
}

func getClientConn(ctx *cli.Context, skipMacaroons bool) *grpc.ClientConn {
	// Create a dial options array.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(ctx.GlobalString("rpcserver"), opts...)
	if err != nil {
		fatal(err)
	}

	return conn
}

func main() {
	app := cli.NewApp()
	app.Name = "pscli"
	app.Version = fmt.Sprintf("0.1")
	app.Usage = "Control plane for your PayServer Daemon (psd)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpcserver",
			Value: defaultRPCHostPort,
			Usage: "host:port of payserver",
		},
	}
	app.Commands = []cli.Command{
		createReceiptCommand,
		validateReceiptCommand,
		balanceCommand,
		estimateFeeCommand,
		sendPaymentCommand,
		paymentByIDCommand,
		paymentByReceiptCommand,
		listPaymentsCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
