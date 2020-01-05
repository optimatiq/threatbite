package controllers

import (
	"net"

	lru "github.com/hashicorp/golang-lru"
	"github.com/optimatiq/threatbite/ip"
)

// IPResult response object, which contains detailed information returned from Check method.
type IPResult struct {
	Scoring       uint8  `json:"scoring"`
	Action        string `json:"action"`
	Company       string `json:"company"`
	Country       string `json:"country"`
	BadReputation bool   `json:"bad"`
	Bot           bool   `json:"bot"`
	Datacenter    bool   `json:"dc"`
	Private       bool   `json:"private"`
	Proxy         bool   `json:"proxy"`
	SearchEngine  bool   `json:"se"`
	Spam          bool   `json:"spam"`
	Tor           bool   `json:"tor"`
	Vpn           bool   `json:"vpn"`
}

// IP a container for IP controller.
type IP struct {
	ipinfo *ip.IP
	cache  *lru.Cache
}

// NewIP creates new IP scoring controller.
func NewIP(ipinfo *ip.IP) (*IP, error) {
	cache, err := lru.New(4096)
	if err != nil {
		return nil, err
	}

	ip := &IP{
		cache:  cache,
		ipinfo: ipinfo,
	}

	return ip, nil
}

// Validate returns nil if provided IP address is valid,
// otherwise, returns error, which can be presented to the user.
func (i *IP) Validate(addr string) error {
	if net.ParseIP(addr) == nil {
		return ErrInvalidIP
	}
	return nil
}

// Check is the main module functions, which is used to perform all checks for given argument.
func (i *IP) Check(addr string) (*IPResult, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, ErrInvalidIP
	}

	if v, ok := i.cache.Get(addr); ok {
		return v.(*IPResult), nil
	}

	info, err := i.ipinfo.GetInfo(ip)
	if err != nil {
		return nil, err
	}

	result := &IPResult{
		Scoring:      info.IPScoring,
		Country:      info.Country,
		Company:      info.Company,
		Tor:          info.IsTor,
		Proxy:        info.IsProxy,
		SearchEngine: info.IsSearchEngine,
		Private:      info.IsPrivate,
		Spam:         info.IsSpam,
		Datacenter:   info.IsDatacenter,
		Vpn:          info.IsVpn,
	}

	if !i.cache.Contains(addr) {
		i.cache.Add(addr, result)
	}

	return result, nil
}
