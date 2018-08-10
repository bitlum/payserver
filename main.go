package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"path/filepath"

	"net"
	"sync"

	"github.com/bitlum/connector/connectors/bitcoind"
	rpc "github.com/bitlum/connector/crpc"
	"github.com/bitlum/connector/estimator"
	"github.com/bitlum/connector/connectors/geth"
	"github.com/bitlum/connector/connectors/lnd"
	"github.com/bitlum/connector/metrics"
	cryptoMetrics "github.com/bitlum/connector/metrics/crypto"
	rpcMetrics "github.com/bitlum/connector/metrics/rpc"
	"github.com/bitlum/viabtc_rpc_client"
	"github.com/btcsuite/go-flags"
	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/common"
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

	var client *viabtc.Client
	if !loadedConfig.EngineDisabled {

		// Create client client in order to be able to communicate with exchange
		// client itself.
		mainLog.Infof("Initialize client client %v:%v", loadedConfig.EngineHost,
			loadedConfig.EnginePort)
		client = viabtc.NewClient(&viabtc.Config{
			Host: loadedConfig.EngineHost,
			Port: loadedConfig.EnginePort,
		})
	}

	// TODO(andrew.shvv) add net config and daemon checks
	mainLog.Infof("Initialising metric for crypto clients...")
	cryptoMetricsBackend, err := cryptoMetrics.InitMetricsBackend(loadedConfig.Network)
	if err != nil {
		return errors.Errorf("unable to init bitcoind metrics: %v", err)
	}

	blockchainConnectors := make(map[viabtc.AssetType]connectors.BlockchainConnector)
	lightningConnectors := make(map[viabtc.AssetType]connectors.LightningConnector)

	// Create blockchain connectors in order to be able to listen for incoming
	// transaction, be able to answer on the question how many
	// pending transaction user have and also to withdraw money from exchange.
	if !loadedConfig.BitcoinCash.Disabled {
		bitcoinCashConnector, err := bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.BitcoinCash.MinConfirmations,
			SyncLoopDelay:    loadedConfig.BitcoinCash.SyncDelay,
			DataDir:          loadedConfig.DataDir,
			Asset:            viabtc.AssetBCH,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerUnit: loadedConfig.BitcoinCash.FeePerUnit,
			DaemonCfg: &bitcoind.DaemonConfig{
				Name:       "bitcoinabc",
				ServerHost: loadedConfig.BitcoinCash.Host,
				ServerPort: loadedConfig.BitcoinCash.Port,
				User:       loadedConfig.BitcoinCash.User,
				Password:   loadedConfig.BitcoinCash.Password,
			},
		})
		if err != nil {
			return errors.Errorf("unable to create bitcoin cash connector: %v", err)
		}

		blockchainConnectors[viabtc.AssetBCH] = bitcoinCashConnector
	}

	if !loadedConfig.Bitcoin.Disabled {
		bitcoinConnector, err := bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Bitcoin.MinConfirmations,
			SyncLoopDelay:    loadedConfig.Bitcoin.SyncDelay,
			DataDir:          loadedConfig.DataDir,
			Asset:            viabtc.AssetBTC,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerUnit: loadedConfig.BitcoinCash.FeePerUnit,
			DaemonCfg: &bitcoind.DaemonConfig{
				Name:       "bitcoind",
				ServerHost: loadedConfig.Bitcoin.Host,
				ServerPort: loadedConfig.Bitcoin.Port,
				User:       loadedConfig.Bitcoin.User,
				Password:   loadedConfig.Bitcoin.Password,
			},
		})
		if err != nil {
			return errors.Errorf("unable to create bitcoin connector: %v", err)
		}

		blockchainConnectors[viabtc.AssetBTC] = bitcoinConnector
	}

	if !loadedConfig.Dash.Disabled {
		dashConnector, err := bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Dash.MinConfirmations,
			SyncLoopDelay:    loadedConfig.Dash.SyncDelay,
			DataDir:          loadedConfig.DataDir,
			Asset:            viabtc.AssetDASH,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerUnit: loadedConfig.Dash.FeePerUnit,
			DaemonCfg: &bitcoind.DaemonConfig{
				Name:       "dashd",
				ServerHost: loadedConfig.Dash.Host,
				ServerPort: loadedConfig.Dash.Port,
				User:       loadedConfig.Dash.User,
				Password:   loadedConfig.Dash.Password,
			},
		})
		if err != nil {
			return errors.Errorf("unable to create dash connector: %v", err)
		}

		blockchainConnectors[viabtc.AssetDASH] = dashConnector
	}

	if !loadedConfig.Litecoin.Disabled {
		litecoinConnector, err := bitcoind.NewConnector(&bitcoind.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Litecoin.MinConfirmations,
			SyncLoopDelay:    loadedConfig.Litecoin.SyncDelay,
			DataDir:          loadedConfig.DataDir,
			Asset:            viabtc.AssetLTC,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
			// TODO(andrew.shvv) Create subsystem to return current fee per unit
			FeePerUnit: loadedConfig.Litecoin.FeePerUnit,
			DaemonCfg: &bitcoind.DaemonConfig{
				Name:       "litecoind",
				ServerHost: loadedConfig.Litecoin.Host,
				ServerPort: loadedConfig.Litecoin.Port,
				User:       loadedConfig.Litecoin.User,
				Password:   loadedConfig.Litecoin.Password,
			},
		})
		if err != nil {
			return errors.Errorf("unable to create litecoin connector: %v", err)
		}

		blockchainConnectors[viabtc.AssetLTC] = litecoinConnector
	}

	if !loadedConfig.Ethereum.Disabled {
		ethConnector, err := geth.NewConnector(&geth.Config{
			Net:              loadedConfig.Network,
			MinConfirmations: loadedConfig.Ethereum.MinConfirmations,
			SyncTickDelay:    loadedConfig.Ethereum.SyncDelay,
			DataDir:          loadedConfig.DataDir,
			Asset:            viabtc.AssetETH,
			Logger:           mainLog,
			Metrics:          cryptoMetricsBackend,
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

		blockchainConnectors[viabtc.AssetETH] = ethConnector
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
		})
		if err != nil {
			return errors.Errorf("unable to create lightning bitcoin connector"+
				": %v", err)
		}

		if err := lightningConnector.Start(); err != nil {
			return errors.Errorf("unable to create lightning bitcoin client: %v",
				err)
		}
		defer func() {
			if err := lightningConnector.Stop("stopped by user"); err != nil {
				mainLog.Warn("unable to shutdown lightning bitcoin"+
					" connector: %v", err)
			}
		}()

		lightningConnectors[viabtc.AssetBTC] = lightningConnector
	}

	for asset, connector := range blockchainConnectors {
		switch c := connector.(type) {
		case *bitcoind.Connector:
			if err := c.Start(); err != nil {
				return errors.Errorf("unable to start %v connector: %v",
					asset, err)
			}
		case *geth.Connector:
			if err := c.Start(); err != nil {
				return errors.Errorf("unable to start %v connector: %v",
					asset, err)
			}
		}
	}

	estmtr := estimator.NewCoinmarketcapEstimator()
	if err := estmtr.Start(); err != nil {
		return errors.Errorf("unable to start estimator: %v", err)
	}

	// Initialise the metric endpoint. This endpoint is used by the metric
	// server to collect the metric from.
	metricsEndpointAddr := net.JoinHostPort(loadedConfig.Prometheus.Host,
		loadedConfig.Prometheus.Port)
	metrics.StartServer(metricsEndpointAddr)

	// TODO(andrew.shvv) add net config and daemon checks
	mainLog.Infof("Initialising metric for rpc...")
	rpcMetricsBackend, err := rpcMetrics.InitMetricsBackend(loadedConfig.Network)
	if err != nil {
		return errors.Errorf("unable to init rpc metrics: %v", err)
	}

	paymentsStore := common.NewMemoryPaymentsStore(30 * 24 * time.Hour)
	paymentsStore.StartCleaner()

	// Initialize RPC server to handle gRPC requests from trading bots and
	// frontend users.
	rpcServer, err := rpc.NewRPCServer(loadedConfig.Network, blockchainConnectors,
		lightningConnectors, paymentsStore, estmtr, rpcMetricsBackend)
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

	quit := make(chan struct{})
	var wg sync.WaitGroup

	if blockchainConnectors != nil {
		for asset, connector := range blockchainConnectors {
			mainLog.Infof("Initialize blockchain connector for '%v' asset",
				asset)

			wg.Add(1)
			go func(asset viabtc.AssetType, connector connectors.BlockchainConnector) {
				defer wg.Done()

				for {
					select {
					case <-quit:
						return
					case payments := <-connector.ReceivedPayments():
						for _, payment := range payments {
							if !loadedConfig.EngineDisabled {
								doDeposit(client, payment, asset)
							}
							paymentsStore.AddPayment(payment)
						}
					}
				}
			}(asset, connector)
		}
	} else {
		mainLog.Warnf("connector client haven't been initialized, " +
			"skipping running the transaction notification listener")
	}

	if lightningConnectors != nil {
		for asset, connector := range lightningConnectors {
			mainLog.Infof("Initialize lightning connector for '%v' asset",
				asset)

			wg.Add(1)
			go func(asset viabtc.AssetType, connector connectors.LightningConnector) {
				defer wg.Done()

				for {
					select {
					case <-quit:
						return
					case payment := <-connector.ReceivedPayments():
						if !loadedConfig.EngineDisabled {
							doDeposit(client, payment, asset)
						}
						paymentsStore.AddPayment(payment)
					}
				}
			}(asset, connector)
		}
	} else {
		mainLog.Warnf("connector client haven't been initialized, " +
			"skipping running the transaction notification listener")
	}

	addInterruptHandler(shutdownChannel, func() {
		paymentsStore.StopCleaner()
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
		estmtr.Stop()
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
