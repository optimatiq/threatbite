package datasource

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/asergeyev/nradix"
	"github.com/labstack/gommon/log"
	"github.com/patrickmn/go-cache"
)

// IPNet container struct for IP/CIDR operations
type IPNet struct {
	cache     *cache.Cache
	cidrs     *nradix.Tree
	cidrsLock sync.RWMutex
	ips       map[uint64]bool
	ipsLock   sync.RWMutex
	ds        DataSource
	name      string
}

// NewIPNet returns a new IP/CIDR list build on top of radix tree (for CIDRS) and go map for IPs.
func NewIPNet(ds DataSource, name string) *IPNet {
	return &IPNet{
		cache: cache.New(1*time.Minute, 1*time.Minute),
		cidrs: nradix.NewTree(0),
		ips:   make(map[uint64]bool),
		ds:    ds,
		name:  name,
	}
}

// Check if lists contains IP from request
func (l *IPNet) Check(ip net.IP) (bool, error) {
	ipString := ip.String()
	keyPermBlockIP := "ip_" + ipString
	if _, ok := l.cache.Get(keyPermBlockIP); ok {
		return true, nil
	}

	l.cidrsLock.RLock()
	cidrFound, err := l.cidrs.FindCIDR(ipString)
	l.cidrsLock.RUnlock()
	if err != nil {
		return false, fmt.Errorf("could not find element: %s, error: %w", ipString, err)
	}

	if cidrFound != nil {
		l.cache.Set(keyPermBlockIP, cidrFound, 0)
		return true, nil
	}

	uip, _ := l.ipToUint64(ip)
	l.ipsLock.RLock()
	value, ipFound := l.ips[uip]
	l.ipsLock.RUnlock()
	if ipFound {
		l.cache.Set(keyPermBlockIP, value, 0)
		return true, nil
	}

	return false, nil
}

// Close clears underlying radix tree.
func (l *IPNet) Close() {
	l.ipsLock.Lock()
	l.ips = map[uint64]bool{}
	l.ipsLock.Unlock()

	l.cidrsLock.Lock()
	l.cidrs = nradix.NewTree(0)
	l.cidrsLock.Unlock()

	l.cache.Flush()
}
func (l *IPNet) Load() error {
	var cidrs int

	cidrsTemp := nradix.NewTree(0)
	ipsTemp := map[uint64]bool{}

	log.Debugf("[list] loading %s list start", l.name)
	defer func() {
		l.ipsLock.Lock()
		l.ips = ipsTemp
		l.ipsLock.Unlock()

		l.cidrsLock.Lock()
		l.cidrs = cidrsTemp
		l.cidrsLock.Unlock()

		l.cache.Flush()

		log.Debugf("[list] loading %s stop; stats IPs: %d, CIDRs: %d", l.name, len(l.ips), cidrs)
	}()

	if err := l.ds.Reset(); err != nil {
		return fmt.Errorf("could not reset data source, error: %w", err)
	}

	for {
		ipNet, err := l.ds.Next()
		if err != nil {
			if err == ErrNoData {
				return nil
			} else if err == ErrInvalidData {
				continue
			} else {
				return fmt.Errorf("could not iterate over data source, error: %w", err)
			}
		}

		// single IP address, not a CIDR, mask contains only "ones"
		if ones, bits := ipNet.Mask.Size(); ones == bits {
			i, _ := l.ipToUint64(ipNet.IP)
			ipsTemp[i] = true
			continue
		}

		err = cidrsTemp.AddCIDR(ipNet.String(), true)
		if err != nil && err != nradix.ErrNodeBusy {
			return fmt.Errorf("could not add IP: %s, error: %w", ipNet.String(), err)
		}
		cidrs++
	}
}

func (l *IPNet) ipToUint64(ip net.IP) (uint64, error) {
	if to4 := ip.To4(); to4 != nil {
		return uint64(to4[0])<<24 | uint64(to4[1])<<16 | uint64(to4[2])<<8 | uint64(to4[3]), nil
	} else if to16 := ip.To16(); to16 != nil {
		int1 := uint64(0)
		for i := 0; i < 8; i++ {
			int1 = (int1 << 8) + uint64(ip[i])
		}
		return int1, nil
	}
	return 0, errors.New("could not convert IP address")
}
