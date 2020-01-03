package datasource

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

// ListDataSource stores current state (counters) of this source.
type ListDataSource struct {
	i     int
	iLock sync.Mutex
	ips   []*net.IPNet
}

// NewListDataSource first argument is a list of IPs or CIDRs.
// Returns DataSource or error on IP parsing.
func NewListDataSource(list []string) (*ListDataSource, error) {
	var ipNets []*net.IPNet
	for _, element := range list {
		if strings.Contains(element, "/") {
			_, ipNet, err := net.ParseCIDR(element)
			if err != nil {
				return nil, fmt.Errorf("invalid address: %s, error: %w", element, err)
			}

			ipNets = append(ipNets, ipNet)
		} else {
			ip := net.ParseIP(element)
			if ip == nil {
				return nil, errors.New("invalid address: " + element)
			}

			ipNets = append(ipNets, &net.IPNet{IP: ip, Mask: net.CIDRMask(8*len(ip), 8*len(ip))})
		}
	}
	return &ListDataSource{
		ips: ipNets,
	}, nil
}

// Reset rewinds source to the beginning.
func (s *ListDataSource) Reset() error {
	s.iLock.Lock()
	defer s.iLock.Unlock()

	s.i = 0
	return nil
}

// Next returns IP/CIDR from the provided list in NewListDataSource method.
// ErrNoData is returned when there is no data, this error indicates that we reached the end.
func (s *ListDataSource) Next() (*net.IPNet, error) {
	s.iLock.Lock()
	defer s.iLock.Unlock()

	if s.i >= len(s.ips) || len(s.ips) <= 0 {
		return nil, ErrNoData
	}

	v := s.ips[s.i]
	s.i++
	return v, nil
}
