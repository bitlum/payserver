package rpc

import (
	"testing"
	"time"
	"github.com/bitlum/connector/metrics"
	"github.com/bitlum/connector/crpc/go"
)

func TestGraphQLMetrics(t *testing.T) {
	crypto.StartServer("localhost:9999")

	m, err := InitMetricsBackend("simnet")
	if err != nil {
		t.Fatalf("want no error but got `%v`", err)
	}

	for {
		m.AddError(crpc.CreateAddress, crpc.ErrNetworkNotSupported)
		time.Sleep(time.Second)
	}
}
