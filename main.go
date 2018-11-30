package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"net"
	"sync"

	"github.com/bitlum/connector/connectors"
	bitcoind "github.com/bitlum/connector/connectors/daemons/bitcoind_simple"
	"github.com/bitlum/connector/connectors/daemons/geth"
	"github.com/bitlum/connector/connectors/daemons/lnd"
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/bitlum/connector/connectors/rpc/bitcoincash"
	"github.com/bitlum/connector/connectors/rpc/dash"
	"github.com/bitlum/connector/connectors/rpc/litecoin"
	rpc "github.com/bitlum/connector/crpc"
	"github.com/bitlum/connector/db/sqlite"
	"github.com/bitlum/connector/metrics"
	cryptoMetrics "github.com/bitlum/connector/metrics/crypto"
	rpcMetrics "github.com/bitlum/connector/metrics/rpc"
	"github.com/btcsuite/go-flags"
	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

var (
	// shutdownChannel is used to identify that process creator send us signal to
	// shutdown the backend service.
	shutdownChannel = make(chan struct{})
)

func backendMain() error {
	// Load the configuration, and parse any command line options.
	defaultConfig := getDefaultConfig()
	if err := defaultConfig.loadConfig(); err != nil {
		return err
	}
	loadedConfig := defaultConfig

	logFile := filepath.Join(loadedConfig.LogDir, defaultLogFilename)
	closeRotator := initLogRotator(logFile)
	defer closeRotator()

	mainLog.Infof("Initialising metric for crypto clients...")
	cryptoMetricsBackend, err := cryptoMetrics.InitMetricsBackend(loadedConfig.Network)
	if err != nil {
		return errors.Errorf("unable to init bitcoind metrics: %v", err)
	}

	mainLog.Infof("Initialising metric for rpc...")
	rpcMetricsBackend, err := rpcMetrics.InitMetricsBackend(loadedConfig.Network)
	if err != nil {
		return errors.Errorf("unable to init rpc metrics: %v", err)
	}

	// Channel is used to notify spawned in the main func goroutines that
	// daemon is shutting down.
	quit := make(chan struct{}, 0)

	blockchainConnectors := make(map[connectors.Asset]connectors.BlockchainConnector)
	lightningConnectors := make(map[connectors.Asset]connectors.LightningConnector)

	db, err := sqlite.Open(loadedConfig.DataDir, "sqlite")
	if err != nil {
		return errors.Errorf("unable open sqlite db: %v", err)
	}

	if err := db.Migrate(); err != nil {
		return errors.Errorf("unable to migrate: %v", err)
	}

	paymentsStore := &sqlite.PaymentsStore{DB: db}

	bitcoinRPCClient, err := bitcoin.NewClient(bitcoin.ClientConfig{
		Name:     "bitcoind",
		Logger:   rpcLog,
		Asset:    connectors.BTC,
		RPCHost:  loadedConfig.Bitcoin.Host,
		RPCPort:  loadedConfig.Bitcoin.Port,
		User:     loadedConfig.Bitcoin.User,
		Password: loadedConfig.Bitcoin.Password,
	})
	if err != nil {
		return errors.Errorf("unable to create bitcoin rpc client: %v")
	}

	bitcoincashRPCClient, err := bitcoincash.NewClient(bitcoincash.ClientConfig{
		Name:     "bitcoinabc",
		Logger:   rpcLog,
		Asset:    connectors.BCH,
		RPCHost:  loadedConfig.BitcoinCash.Host,
		RPCPort:  loadedConfig.BitcoinCash.Port,
		User:     loadedConfig.BitcoinCash.User,
		Password: loadedConfig.BitcoinCash.Password,
	})
	if err != nil {
		return errors.Errorf("unable to create bitcoin cash rpc client: %v")
	}

	dashRPCClient, err := dash.NewClient(dash.ClientConfig{
		Name:     "dashd",
		Logger:   rpcLog,
		Asset:    connectors.DASH,
		RPCHost:  loadedConfig.Dash.Host,
		RPCPort:  loadedConfig.Dash.Port,
		User:     loadedConfig.Dash.User,
		Password: loadedConfig.Dash.Password,
	})
	if err != nil {
		return errors.Errorf("unable to create dash rpc client: %v")
	}

	litecoinRPCClient, err := litecoin.NewClient(litecoin.ClientConfig{
		Name:     "litecoind",
		Logger:   rpcLog,
		Asset:    connectors.LTC,
		RPCHost:  loadedConfig.Litecoin.Host,
		RPCPort:  loadedConfig.Litecoin.Port,
		User:     loadedConfig.Litecoin.User,
		Password: loadedConfig.Litecoin.Password,
	})
	if err != nil {
		return errors.Errorf("unable to create dash rpc client: %v")
	}

	// Create blockchain connectors in order to be able to listen for incoming
	// transaction, be able to answer on the question how many
	// pending transaction user have and also to withdraw money from exchange.
	if !loadedConfig.BitcoinCash.Disabled {
		blockchainConnectors[connectors.BCH], err = bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.BitcoinCash.MinConfirmations,
			Asset:            connectors.BCH,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			PaymentStore:     paymentsStore,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerByte: loadedConfig.BitcoinCash.FeePerUnit,
			RPCClient:  bitcoincashRPCClient,
		})
		if err != nil {
			return errors.Errorf("unable to create bitcoin cash connector: %v", err)
		}
	}

	if !loadedConfig.Bitcoin.Disabled {
		blockchainConnectors[connectors.BTC], err = bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Bitcoin.MinConfirmations,
			Asset:            connectors.BTC,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			PaymentStore:     paymentsStore,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerByte: loadedConfig.BitcoinCash.FeePerUnit,
			RPCClient:  bitcoinRPCClient,
		})
		if err != nil {
			return errors.Errorf("unable to create bitcoin connector: %v", err)
		}
	}

	if !loadedConfig.Dash.Disabled {
		blockchainConnectors[connectors.DASH], err = bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Dash.MinConfirmations,
			Asset:            connectors.DASH,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			PaymentStore:     paymentsStore,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerByte: loadedConfig.Dash.FeePerUnit,
			RPCClient:  dashRPCClient,
		})
		if err != nil {
			return errors.Errorf("unable to create dash connector: %v", err)
		}
	}

	if !loadedConfig.Litecoin.Disabled {
		blockchainConnectors[connectors.LTC], err = bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Litecoin.MinConfirmations,
			Asset:            connectors.LTC,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			PaymentStore:     paymentsStore,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerByte: loadedConfig.Litecoin.FeePerUnit,
			RPCClient:  litecoinRPCClient,
		})
		if err != nil {
			return errors.Errorf("unable to create litecoin connector: %v", err)
		}
	}

	if !loadedConfig.Ethereum.Disabled {
		blockchainConnectors[connectors.ETH], err = geth.NewConnector(&geth.Config{
			Net:                 loadedConfig.Network,
			MinConfirmations:    loadedConfig.Ethereum.MinConfirmations,
			SyncTickDelay:       loadedConfig.Ethereum.SyncDelay,
			Asset:               connectors.ETH,
			Logger:              mainLog,
			Metrics:             cryptoMetricsBackend,
			LastSyncedBlockHash: loadedConfig.Ethereum.ForceLastHash,
			PaymentStorage:      paymentsStore,
			StateStorage:        sqlite.NewConnectorStateStorage(connectors.ETH, db),
			AccountStorage:      sqlite.NewGethAccountsStorage(db),
			DaemonCfg: &geth.DaemonConfig{
				Name:       "geth",
				ServerHost: loadedConfig.Ethereum.Host,
				ServerPort: loadedConfig.Ethereum.Port,
				Password:   loadedConfig.Ethereum.Password,
			},
		})
		if err != nil {
			return errors.Errorf("unable to create ethereum connector: %v", err)
		}
	}

	if !loadedConfig.BitcoinLightning.Disabled {
		lightningConnector, err := lnd.NewConnector(&lnd.Config{
			PeerHost:     loadedConfig.BitcoinLightning.PeerHost,
			PeerPort:     loadedConfig.BitcoinLightning.PeerPort,
			Net:          loadedConfig.Network,
			Name:         "lnd",
			Host:         loadedConfig.BitcoinLightning.Host,
			Port:         loadedConfig.BitcoinLightning.Port,
			TlsCertPath:  loadedConfig.BitcoinLightning.TlsCertPath,
			MacaroonPath: loadedConfig.BitcoinLightning.MacaroonPath,
			Metrics:      cryptoMetricsBackend,
			PaymentStore: paymentsStore,
		})
		if err != nil {
			return errors.Errorf("unable to create lightning bitcoin "+
				"connector: %v", err)
		}

		// Retry start connector until daemon will exit or connector start
		// succeed. It is needed so that prometheus could scratch the fail
		// start metric and send alert.
		go func(c *lnd.Connector) {
			for {
				if err := c.Start(); err != nil {
					mainLog.Errorf("unable to start BTC lightning "+
						" connector: %v", err)

					select {
					case <-time.After(5 * time.Second):
						mainLog.Infof("Retrying start BTC lightning connector")
						continue
					case <-quit:
						return
					}
				}

				return
			}
		}(lightningConnector)

		defer func() {
			if err := lightningConnector.Stop("stopped by user"); err != nil {
				mainLog.Warn("unable to shutdown lightning bitcoin"+
					" connector: %v", err)
			}
		}()

		lightningConnectors[connectors.BTC] = lightningConnector
	}

	for asset, connector := range blockchainConnectors {
		switch c := connector.(type) {
		case *bitcoind.Connector:
			// Retry start connector until daemon will exit or connector start
			// succeed. It is needed so that prometheus could scratch the fail
			// start metric and send alert.
			go func(c *bitcoind.Connector, asset connectors.Asset) {
				for {
					if err := c.Start(); err != nil {
						mainLog.Errorf("unable to start %v blockchain "+
							"connector: %v", asset, err)

						select {
						case <-time.After(5 * time.Second):
							mainLog.Infof("Retrying start %v connector", asset)
							continue
						case <-quit:
							return
						}
					}

					return
				}
			}(c, asset)

		case *geth.Connector:
			// Retry start connector until daemon will exit or connector start
			// succeed. It is needed so that prometheus could scratch the fail
			// start metric and send alert.
			go func(c *geth.Connector, asset connectors.Asset) {
				for {
					if err := c.Start(); err != nil {
						mainLog.Errorf("unable to start %v connector: %v",
							asset, err)

						select {
						case <-time.After(5 * time.Second):
							mainLog.Infof("Retrying start lightning ETH " +
								"connector")
							continue
						case <-quit:
							return
						}
					}

					return
				}
			}(c, asset)
		}
	}

	// Initialise the metric endpoint. This endpoint is used by the metric
	// server to collect the metric from.
	metricsEndpointAddr := net.JoinHostPort(loadedConfig.Prometheus.Host,
		loadedConfig.Prometheus.Port)
	metrics.StartServer(metricsEndpointAddr)

	// Initialize RPC server to handle gRPC requests from trading bots and
	// frontend users.
	rpcServer, err := rpc.NewRPCServer(loadedConfig.Network, blockchainConnectors,
		lightningConnectors, paymentsStore, rpcMetricsBackend)
	if err != nil {
		return errors.Errorf("unable to init RPC server: %v", err)
	}

	var opts []grpc.ServerOption

	// If TLS files are exist than use it to encrypt gRPC endpoints
	// communications.
	if fileExists(loadedConfig.TLSCertPath) && fileExists(loadedConfig.TLSKeyPath) {
		creds, err := credentials.NewServerTLSFromFile(loadedConfig.TLSCertPath,
			loadedConfig.TLSKeyPath)
		if err != nil {
			return errors.Errorf("unable to load TLS keys: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		mainLog.Info("TLS encryption enabled")
	}

	grpcServer := grpc.NewServer(opts...)
	rpc.RegisterPayServerServer(grpcServer, rpcServer)

	grpcAddr := net.JoinHostPort(loadedConfig.RPCHost, loadedConfig.RPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return errors.Errorf("unable to listen on gRPC addr: %v", err)
	}

	// Spawn goroutine which runs the original gRPC server, which will be
	// responsible for transferring requests from trading robots to the rpc
	// server.
	errChan := make(chan error)
	go func() {
		mainLog.Infof("server gRPC on addr: '%v'", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- errors.Errorf("unable to server gRPC server: %v", err)
			return
		}
		mainLog.Info("stop serving gRPC")
	}()

	var wg sync.WaitGroup

	addInterruptHandler(shutdownChannel, func() {
		grpcServer.Stop()

		for _, c := range blockchainConnectors {
			switch c := c.(type) {
			case *bitcoind.Connector:
				c.Stop("stopped by user")
			case *geth.Connector:
				c.Stop("stopped by user")
			}
		}

		close(quit)
		wg.Wait()
	})

	select {
	case <-shutdownChannel:
		mainLog.Info("Shutting down connector")
		return nil
	case err := <-errChan:
		return err
	}
}

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Call the "real" main in a nested manner so the defers will properly
	// be executed in the case of a graceful shutdown.
	if err := backendMain(); err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
