package utils

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

func GetNetworkConnection(onionService string, port int, proxyAddress string, timeout time.Duration) (net.Conn, error) {
	portNumber := strconv.Itoa(port)

	torDialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	start := time.Now()
	conn, err := torDialer.(proxy.ContextDialer).DialContext(ctx, "tcp", onionService+":"+portNumber)
	// conn, err := torDialer.Dial("tcp", onionService+":"+portNumber)
	end := time.Now()

	log.Printf("Time difference just for dialing tcp connection is: %s", end.Sub(start))

	if err != nil {
		return nil, err
	}

	// conn.SetDeadline(time.Now().Add(timeout * time.Second))
	return conn, err
}
