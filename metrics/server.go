package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"time"
)

const (
	// Namespace is used to distinguish this metrics from other. Namespace is
	// global for the project, for that reason ensure that name not interfere
	// with other namespaces on your prometheus server.
	Namespace = "connector"

	// NetLabel is used to distinguish different blockchain network names in
	// which service is working e.g simnet, testnet, mainnet, during the
	// process of metric analysis and alert rule constructing.
	NetLabel = "net"
)

func StartServer(addr string) *http.Server {

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())

	server := &http.Server{Addr: addr, Handler: handler}

	go func() {
		log.Infof("Starting metrics http server on `%s`", addr)

		for {
			switch err := server.ListenAndServe(); err {
			case http.ErrServerClosed:
				log.Infof("Metrics http server shutdown")
				return
			default:
				log.Errorf("Metrics http server error: %v", err)

				time.Sleep(5 * time.Second)

				log.Infof("Trying to start metrics http"+
					" server on '%s' one more time", addr)
			}
		}
	}()

	return server
}
