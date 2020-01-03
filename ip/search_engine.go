package ip

import (
	"net"
	"regexp"
	"time"

	"github.com/labstack/gommon/log"
)

type searchEngine struct {
	geoip *maxmind
}

func newSearchEngine(geoip *maxmind) *searchEngine {
	return &searchEngine{geoip: geoip}
}

var searchHosts = regexp.MustCompile("googlebot.com|google.com|yandex.com|search.msn.com|yahoo.net|yahoo.com|yahoo-net.jp|yahoo.co.jp|crawl.baidu.com|opera-mini.net|seznam.cz|mail.ru|pinterest.com|archive.org")
var searchASNs = regexp.MustCompile("Google|Seznam.cz|Microsoft|Yahoo|Yandex|Opera Software|Facebook|Mail.Ru|Apple|LinkedIn|Twitter Inc.|Internet Archive")

func (s *searchEngine) isSearchEngine(ip net.IP) (bool, error) {
	asn, err := s.geoip.getCompany(ip)
	if err != nil {
		return false, err
	}

	if searchASNs.MatchString(asn) {
		log.Debugf("[isEngine] ip: %s Company: %s %t", ip, asn, true)
		return true, nil
	}

	hostnames, err := lookupAddrWithTimeout(ip.String(), 500*time.Millisecond)
	if err != nil {
		// errors like "no such host" are normal, we don't need to pollute error logs
		log.Debugf("[isEngine] ip: %s error: %s", ip, err)
		return false, nil
	}
	ips, err := lookupIPWithTimeout(hostnames[0], 500*time.Millisecond)
	if err != nil {
		// errors like "cannot lookup" are normal, we don't need to pollute error logs
		log.Debugf("[isEngine] ip: %s error: %s", ip, err)
		return false, nil
	}

	matchedIP := false
	for _, i := range ips {
		if i.Equal(ip) {
			matchedIP = true
			break
		}
	}
	if !matchedIP {
		log.Debugf("[isEngine] ip: %s and hosts: %v don't match", ip, hostnames)
		return false, nil
	}

	for _, h := range hostnames {
		if searchHosts.MatchString(h) {
			log.Debugf("[isEngine] ip: %s Company: %s %t", ip, asn, true)
			return true, nil
		}
	}

	return false, nil
}
