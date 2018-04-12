package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"log"

	"github.com/bitlum/btcutil"
	"github.com/btcsuite/go-flags"
)

const (
	defaultRPCHost = "0.0.0.0"
	defaultRPCPort = "9002"

	defaultPrometheusEndpointHost = "0.0.0.0"
	defaultPrometheusEndpointPort = "9999"

	defaultTLSCertFilename = "server.cert"
	defaultTLSKeyFilename  = "server.key"

	defaultLogDirname  = "logs"
	defaultLogFilename = "connector.log"
	defaultLogLevel    = "info"

	defaultNet = "simnet"

	defaultConfigFilename = "connector.conf"
)

var (
	homeDir            = btcutil.AppDataDir("connector", false)
	defaultConfigFile  = filepath.Join(homeDir, defaultConfigFilename)
	defaultTLSCertPath = filepath.Join(homeDir, defaultTLSCertFilename)
	defaultTLSKeyPath  = filepath.Join(homeDir, defaultTLSKeyFilename)
	defaultLogDir      = filepath.Join(homeDir, defaultLogDirname)
)

type prometheusConfig struct {
	Host string `long:"host" description:"The host of the prometheus metrics endpoint, from which metric server is trying to fetch metrics"`
	Port string `long:"port" description:"The port of the prometheus metrics endpoint, from which metric server is trying to fetch metrics"`
}

// config defines the configuration options for lnd.
//
// See loadConfig for further details regarding the configuration
// loading+parsing process.
type config struct {
	ShowVersion bool `long:"version" description:"Display version information and exit"`

	TLSCertPath string `long:"tlscertpath" description:"Path to TLS certificate which is used to encrypt RPC endpoint"`
	TLSKeyPath  string `long:"tlskeypath" description:"Path to TLS private key which is used to encrypt RPC endpoint"`

	EngineHost string `long:"enginehost" description:"The host of the exchange engine server"`
	EnginePort int    `long:"engineport" description:"The port of the exchange engine server"`

	RPCHost string `long:"rpchost" description:"The host of the RPC endpoint"`
	RPCPort string `long:"rpcport" description:"The port of the RPC endpoint"`

	Net string `long:"net" description:"The network of the daemon to which connector is connecting" choice:"simnet" choice:"testnet" choice:"mainnet"`

	ConfigFile string `long:"config" description:"Path to configuration file"`

	LogDir     string `long:"logdir" description:"Directory to log output."`
	DebugLevel string `long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`

	Prometheus *prometheusConfig `group:"Prometheus" namespace:"prometheus"`

	Bitcoin          *BitcoindConfig `group:"bitcoin" namespace:"bitcoin"`
	BitcoinLightning *LndConfig      `group:"bitcoinlightning" namespace:"bitcoinlightning"`
	BitcoinCash      *BitcoindConfig `group:"bitcoincash" namespace:"bitcoincash"`
	Litecoin         *BitcoindConfig `group:"litecoin" namespace:"litecoin"`
	Dash             *BitcoindConfig `group:"dash" namespace:"dash"`
	Ethereum         *GethConfig     `group:"ethereum" namespace:"ethereum"`

	DataDir string `long:"datadir" description:"Path to data directory"`
}

type LndConfig struct {
	TlsCertPath string `long:"tlscertpath" description:"Path to the TLS certificate of the lnd daemon"`
	Host        string `long:"host" description:"The host of the lnd daemon"`
	Port        int    `long:"port" description:"The port of the lnd daemon"`

	// TODO(andrew.shvv) Remove when lnd would return this info
	PeerPort string `long:"peerport" description:"Public port of the lnd via which other lightning network nodes could connect"`

	// TODO(andrew.shvv) Remove when lnd would return this info
	PeerHost string `long:"peerhost" description:"Public host of the lnd via which other lightning network nodes could connect"`
}

type GethConfig struct {
	MinConfirmations int    `long:"minconfirmations" description:"Minimum number of block on top of the one where transaction appeared, before we consider transaction as confirmed."`
	SyncDelay        int    `long:"syncdelay" description:"For how long processing loop should sleep before start syncing pending, confirmed and mempool transactions."`
	Host             string `long:"host" description:"The host of the lnd daemon"`
	Port             int    `long:"port" description:"The port of the lnd daemon"`
	User             string `long:"user" description:"Part of the credential information needed to connect to the daemon RPC endpoint"`
	Password         string `long:"password" description:"Part of the credential information needed to connect to the daemon RPC endpoint"`
}

type BitcoindConfig struct {
	MinConfirmations int    `long:"minconfirmations" description:"Minimum number of block on top of the one where transaction appeared, before we consider transaction as confirmed."`
	SyncDelay        int    `long:"syncdelay" description:"For how long processing loop should sleep before start syncing pending, confirmed and mempool transactions."`
	FeePerUnit       int    `long:"feeperunit" description:"Fee for every unit of information needed to put it in the blockchain"`
	Host             string `long:"host" description:"The host of the lnd daemon"`
	Port             int    `long:"port" description:"The port of the lnd daemon"`
	User             string `long:"user" description:"Part of the credential information needed to connect to the daemon RPC endpoint"`
	Password         string `long:"password" description:"Part of the credential information needed to connect to the daemon RPC endpoint"`
}

// getDefaultConfig return default version of service config.
func getDefaultConfig() config {
	return config{
		TLSCertPath: defaultTLSCertPath,
		TLSKeyPath:  defaultTLSKeyPath,

		RPCHost: defaultRPCHost,
		RPCPort: defaultRPCPort,

		ConfigFile: defaultConfigFile,
		LogDir:     defaultLogDir,
		DebugLevel: defaultLogLevel,

		Net: defaultNet,

		Prometheus: &prometheusConfig{
			Host: defaultPrometheusEndpointHost,
			Port: defaultPrometheusEndpointPort,
		},
	}
}

// loadConfig initializes and parses the config using a config file and command
// line options.
//
// The configuration proceeds as follows:
// 	1) Start with a default config with sane settings
// 	2) Pre-parse the command line to check for an alternative config file
// 	3) Load configuration file overwriting defaults with any specified options
// 	4) Parse CLI options and overwrite/add any specified options
func (c *config) loadConfig() error {
	// Pre-parse the command line options to pick up an alternative config
	// file.
	preCfg := c
	if _, err := flags.Parse(preCfg); err != nil {
		return err
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Println(appName, "version", version())
		os.Exit(0)
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	if err := os.MkdirAll(homeDir, 0700); err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				str := "is symlink %s -> %s mounted?"
				err = fmt.Errorf(str, e.Path, link)
			}
		}

		str := "%s: Failed to create home directory: %v"
		err := fmt.Errorf(str, funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// Next, load any additional configuration options from the file.
	var configFileError error
	if err := flags.IniParse(preCfg.ConfigFile, c); err != nil {
		configFileError = err
	}

	// Finally, parse the remaining command line options again to ensure
	// they take precedence.
	if _, err := flags.Parse(c); err != nil {
		return err
	}

	// Ensure that the paths are expanded and cleaned.
	c.TLSCertPath = cleanAndExpandPath(c.TLSCertPath)
	c.TLSKeyPath = cleanAndExpandPath(c.TLSKeyPath)
	c.LogDir = cleanAndExpandPath(c.LogDir)

	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(c.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		return err
	}

	// Warn about missing config file only after all other configuration is
	// done.  This prevents the warning on help messages and invalid
	// options.  Note this should go directly before the return.
	if configFileError != nil {
		log.Printf("%v", configFileError)
	}

	return nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(homeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly. An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(debugLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}

		// Change the logging level for all subsystems.
		setLogLevels(debugLevel)

		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "The specified debug level contains an invalid " +
				"subsystem/level pair [%v]"
			return fmt.Errorf(str, logLevelPair)
		}

		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]

		// Validate subsystem.
		if _, exists := subsystemLoggers[subsysID]; !exists {
			str := "The specified subsystem [%v] is invalid -- " +
				"supported subsytems %v"
			return fmt.Errorf(str, subsysID, supportedSubsystems())
		}

		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, logLevel)
		}

		setLogLevel(subsysID, logLevel)
	}

	return nil
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}

// supportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func supportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}

	// Sort the subsystems for stable display.
	sort.Strings(subsystems)
	return subsystems
}
