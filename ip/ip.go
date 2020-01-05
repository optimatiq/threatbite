package ip

import (
	"context"
	"net"
	"time"

	"github.com/optimatiq/threatbite/ip/datasource"

	"github.com/labstack/gommon/log"
	"golang.org/x/sync/errgroup"
)

// Info a struct, which contains information about IP address.
type Info struct {
	Company        string
	Country        string
	Hostnames      []string
	IsProxy        bool
	IsSearchEngine bool
	IsTor          bool
	IsPrivate      bool
	IsDatacenter   bool
	IsSpam         bool
	IsVpn          bool
	IPScoring      uint8
}

// IP container struct for IP service.
type IP struct {
	tor    *tor
	geoip  geoip
	engine *searchEngine
	proxy  *proxy
	dc     *datacenter
	spam   *spam
	vpn    *vpn
}

// NewIP creates a service for getting information about IP address.
func NewIP(maxmindKey string, proxyDs, spamDs, vpnDs, dcDs datasource.DataSource) *IP {
	geo := newMaxmind(maxmindKey)

	return &IP{
		geoip:  geo,
		tor:    newTor(),
		proxy:  newProxy(proxyDs),
		engine: newSearchEngine(geo),
		dc:     newDC(geo, dcDs),
		spam:   newSpam(spamDs),
		vpn:    newVpn(vpnDs),
	}
}

// GetInfo returns computed information (Info struct) for given IP address.
// Error is returned on critical condition, everything else is logged with debug level.
func (i *IP) GetInfo(ip net.IP) (*Info, error) {
	country, err := i.geoip.getCountry(ip)
	if err != nil {
		return nil, err
	}

	var g errgroup.Group

	var company string
	g.Go(func() (err error) {
		company, err = i.geoip.getCompany(ip)
		return
	})

	var isSearch bool
	g.Go(func() (err error) {
		isSearch, err = i.engine.isSearchEngine(ip)
		return
	})

	var isTor bool
	g.Go(func() (err error) {
		isTor, err = i.tor.isTor(ip)
		return
	})

	var isProxy bool
	g.Go(func() (err error) {
		isProxy, err = i.proxy.isProxy(ip)
		return
	})

	var isDC bool
	g.Go(func() (err error) {
		isDC, err = i.dc.isDC(ip)
		return
	})

	var isSpam bool
	g.Go(func() (err error) {
		isSpam, err = i.spam.isSpam(ip)
		return
	})

	var isVpn bool
	g.Go(func() (err error) {
		isVpn, err = i.vpn.isVpn(ip)
		return
	})

	var isPrivateAddr bool
	g.Go(func() (err error) {
		isPrivateAddr = isPrivateIP(ip)
		return
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// error here can happen, and it's normal
	hostnames, _ := lookupAddrWithTimeout(ip.String(), 500*time.Millisecond)

	// Calculate scoring 0-100 (worst-best)
	var score uint8 = 86

	if isProxy {
		score -= 53
	}
	if !isProxy {
		score += 2
	}

	if isSearch {
		score++
	}

	if isTor {
		score -= 59
	}

	if isDC {
		score -= 16
	}

	if isSpam {
		score -= 24
	}

	if isVpn {
		score -= 13
	}

	if len(hostnames) == 0 {
		score -= 3
	}

	if isPrivateAddr {
		score = 0
	}

	if score > 100 {
		score = 100
	}

	return &Info{
		Company:        company,
		Country:        country,
		IsProxy:        isProxy,
		IsSearchEngine: isSearch,
		IsTor:          isTor,
		Hostnames:      hostnames,
		IsPrivate:      isPrivateAddr,
		IsDatacenter:   isDC,
		IsSpam:         isSpam,
		IsVpn:          isVpn,
		IPScoring:      score,
	}, nil
}

// RunUpdates schedules and runs updates.
// Update interval is defined for each source individually.
func (i *IP) RunUpdates() {
	ctx := context.Background()

	runAndSchedule(ctx, 15*time.Minute, func() {
		if err := i.tor.update(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 24*time.Hour, func() {
		if err := i.geoip.update(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 12*time.Hour, func() {
		if err := i.proxy.ipnet.Load(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 12*time.Hour, func() {
		if err := i.dc.ipnet.Load(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 12*time.Hour, func() {
		if err := i.spam.ipnet.Load(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 12*time.Hour, func() {
		if err := i.vpn.ipnet.Load(); err != nil {
			log.Error(err)
		}
	})
}

func runAndSchedule(ctx context.Context, interval time.Duration, f func()) {
	go func() {
		t := time.NewTimer(0) // first run - immediately
		for {
			select {
			case <-t.C:
				f()
				t = time.NewTimer(interval) // next runs according to the schedule
			case <-ctx.Done():
				return
			}
		}
	}()
}

// isPrivateIP CHeck if IP belongs to private networks
func isPrivateIP(ip net.IP) bool {
	// Eliminate by default multicast and loopback for IPv4 and IPv6
	global := ip.IsGlobalUnicast()
	if !global {
		return true
	}

	privateIPs := []string{
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	}

	for _, i := range privateIPs {
		_, network, _ := net.ParseCIDR(i)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
