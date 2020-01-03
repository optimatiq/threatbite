package ip

import (
	"net"
	"regexp"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/ip/datasource"
)

type proxy struct {
	ipnet *datasource.IPNet
}

func newProxy(source datasource.DataSource) *proxy {
	return &proxy{ipnet: datasource.NewIPNet(source, "proxy")}
}

var reIsProxy = regexp.MustCompile("proxy|sock|anon")

// isProxy check if IP belongs to proxy list or have defined string in reverse name
func (p *proxy) isProxy(ip net.IP) (bool, error) {
	isProxy, err := p.ipnet.Check(ip)
	if isProxy {
		log.Debugf("[isProxy] ip: %s tor: %t", ip, isProxy)
		return isProxy, err
	}

	reverse, err := lookupAddrWithTimeout(ip.String(), 500*time.Millisecond)
	if err != nil {
		// errors like "no such host" are normal, we don't need to pollute error logs
		log.Debugf("[isProxy] ip: %s error: %s", ip, err)
		return false, nil
	}

	if reIsProxy.MatchString(reverse[0]) {
		return true, nil
	}

	return false, nil
}
