package email

import (
	"context"
	"net"
	"time"
)

func lookupMXWithTimeout(name string, timeout time.Duration) ([]*net.MX, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return (&net.Resolver{}).LookupMX(ctx, name)
}

func lookupIPWithTimeout(host string, timeout time.Duration) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	addrs, err := (&net.Resolver{}).LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, ia := range addrs {
		ips[i] = ia.IP
	}
	return ips, nil
}
