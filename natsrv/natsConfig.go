package natsrv

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// StartNats runs an embedded NATS server and returns both the server handle
// (so the caller can shut it down gracefully) and a connected client.
func StartNats() (*server.Server, *nats.Conn, error) {
	ns, err := server.NewServer(&server.Options{
		Port: server.RANDOM_PORT,
	})
	if err != nil {
		return nil, nil, err
	}
	go ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		return nil, nil, fmt.Errorf("nats not ready")
	}
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown()
		return nil, nil, err
	}
	return ns, nc, nil
}
