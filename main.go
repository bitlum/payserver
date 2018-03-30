package main

import (
	"fmt"
	"os"
	"runtime"

	"path/filepath"

	"net"
	"sync"

	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/common/core"
	"github.com/bitlum/connector/bitcoind"
	rpc "github.com/bitlum/connector/crpc/go"
	"github.com/bitlum/connector/estimator"
	"github.com/bitlum/connector/geth"
	"github.com/bitlum/connector/lnd"
	"github.com/btcsuite/go-flags"
	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/bitlum/connector/metrics"
	rpcMetrics "github.com/bitlum/connector/metrics/rpc"
	cryptoMetrics "github.com/bitlum/connector/metrics/crypto"
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

	// Create engine client in order to be able to communicate with exchange
	// engine itself.
	mainLog.Infof("Initialize engine client %v:%v", loadedConfig.EngineHost,
		loadedConfig.EnginePort)
	if err := core.CreateEngine(&core.EngineConfig{
		IP:       loadedConfig.EngineHost,
		HTTPPort: loadedConfig.EnginePort,
	}); err != nil {
		return errors.Errorf("unable to create engine client: %v", err)
	}

	engine, err := core.GetEngine()
	if err != nil {
		return errors.Errorf("unable to get engine: %s", err)
	}

	// TODO(andrew.shvv) add net config and daemon checks
	mainLog.Infof("Initialising metric for crypto clients...")
	cryptoMetricsBackend, err := cryptoMetrics.InitMetricsBackend("simnet")
	if err != nil {
		return errors.Errorf("unable to init bitcoind metrics: %v", err)
	}

	// Create blockchain connectors in order to be able to listen for incoming
	// transaction, be able to answer on the question how many
	// pending transaction user have and also to withdraw money from exchange.
	bitcoinCashConnector, err := bitcoind.NewConnector(&bitcoind.Config{
		MinConfirmations: loadedConfig.BitcoinCash.MinConfirmations,
		SyncLoopDelay:    loadedConfig.BitcoinCash.SyncDelay,
		DataDir:          "/tmp",
		Asset:            core.AssetBCH,
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

	bitcoinConnector, err := bitcoind.NewConnector(&bitcoind.Config{
		MinConfirmations: loadedConfig.Bitcoin.MinConfirmations,
		SyncLoopDelay:    loadedConfig.Bitcoin.SyncDelay,
		DataDir:          "/tmp",
		Asset:            core.AssetBTC,
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

	dashConnector, err := bitcoind.NewConnector(&bitcoind.Config{
		MinConfirmations: loadedConfig.Dash.MinConfirmations,
		SyncLoopDelay:    loadedConfig.Dash.SyncDelay,
		DataDir:          "/tmp",
		Asset:            core.AssetDASH,
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

	litecoinConnector, err := bitcoind.NewConnector(&bitcoind.Config{
		MinConfirmations: loadedConfig.Litecoin.MinConfirmations,
		SyncLoopDelay:    loadedConfig.Litecoin.SyncDelay,
		DataDir:          "/tmp",
		Asset:            core.AssetLTC,
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

	ethConnector, err := geth.NewConnector(&geth.Config{
		MinConfirmations: loadedConfig.Ethereum.MinConfirmations,
		SyncTickDelay:    loadedConfig.Ethereum.SyncDelay,
		DataDir:          "/tmp",
		Asset:            core.AssetETH,
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

	lightningConnector, err := lnd.NewConnector(&lnd.Config{
		Name:        "lnd",
		Host:        loadedConfig.BitcoinLightning.Host,
		Port:        loadedConfig.BitcoinLightning.Port,
		TlsCertPath: loadedConfig.BitcoinLightning.TlsCertPath,
		Metrics:     cryptoMetricsBackend,
	})
	if err != nil {
		return errors.Errorf("unable to create lightning bitcoin connector"+
			": %v", err)
	}

	blockchainConnectors := map[core.AssetType]common.BlockchainConnector{
		core.AssetBTC:  bitcoinConnector,
		core.AssetBCH:  bitcoinCashConnector,
		core.AssetDASH: dashConnector,
		core.AssetLTC:  litecoinConnector,
		core.AssetETH:  ethConnector,
	}

	lightningConnectors := map[core.AssetType]common.LightningConnector{
		core.AssetBTC: lightningConnector,
	}

	for asset, c := range blockchainConnectors {
		switch c := c.(type) {
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

	if err := lightningConnector.Start(); err != nil {
		return errors.Errorf("unable to create lightning bitcoin client: %v",
			err)
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
	rpcMetricsBackend, err := rpcMetrics.InitMetricsBackend("simnet")
	if err != nil {
		return errors.Errorf("unable to init rpc metrics: %v", err)
	}

	// Initialize RPC server to handle gRPC requests from trading bots and
	// frontend users.
	rpcServer, err := rpc.NewRPCServer(blockchainConnectors,
		lightningConnectors, estmtr, mainLog, rpcMetricsBackend)
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
	rpc.RegisterConnectorServer(grpcServer, rpcServer)

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
		for asset, client := range blockchainConnectors {
			mainLog.Infof("Initialize blockchain connector for '%v' asset",
				asset)

			wg.Add(1)
			go func(asset core.AssetType, client common.BlockchainConnector) {
				defer wg.Done()

				for {
					select {
					case <-quit:
						return
					case payments := <-client.ReceivedPayments():
						for _, payment := range payments {
							doDeposit(engine, payment, asset)
						}
					}
				}
			}(asset, client)
		}
	} else {
		mainLog.Warnf("connector client haven't been initialized, " +
			"skipping running the transaction notification listener")
	}

	if lightningConnectors != nil {
		for asset, client := range lightningConnectors {
			mainLog.Infof("Initialize lightning connector for '%v' asset",
				asset)

			wg.Add(1)
			go func(asset core.AssetType, client common.LightningConnector) {
				defer wg.Done()

				for {
					select {
					case <-quit:
						return
					case payment := <-client.ReceivedPayments():
						doDeposit(engine, payment, asset)
					}
				}
			}(asset, client)
		}
	} else {
		mainLog.Warnf("connector client haven't been initialized, " +
			"skipping running the transaction notification listener")
	}

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

		if err := lightningConnector.Stop("stopped by user"); err != nil {
			mainLog.Warn("unable to shutdown lightning bitcoin"+
				" connector: %v", err)
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
