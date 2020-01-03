package ip

import (
	"net"
	"regexp"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/ip/datasource"
)

type vpn struct {
	ipnet *datasource.IPNet
}

func newVpn(source datasource.DataSource) *vpn {
	return &vpn{ipnet: datasource.NewIPNet(source, "vpn")}
}

var reIsVpn = regexp.MustCompile("vpn|ipsec|private|ovudp|l2tp|ovtcp|sstp|expressnetw|anony|hma.rocks|ipvanish|serverlocation.co|world4china|safersoftware.net|dns2use|ivacy|.cstorm.|cryptostorm|boxpnservers|airdns|hide.me|privateinternetaccess|windscribe|lazerpenguin|mullvad")

// isVpn check if IP belongs to vpn list or have defined string in reverse name
func (v *vpn) isVpn(ip net.IP) (bool, error) {
	isVpn, err := v.ipnet.Check(ip)
	if isVpn {
		log.Debugf("[isVpn] ip: %s tor: %t", ip, isVpn)
		return isVpn, err
	}

	reverse, err := lookupAddrWithTimeout(ip.String(), 500*time.Millisecond)
	if err != nil {
		// errors like "no such host" are normal, we don't need to pollute error logs
		log.Debugf("[isVpn] ip: %s error: %s", ip, err)
		return false, nil
	}

	if reIsVpn.MatchString(reverse[0]) {
		return true, nil
	}

	return false, nil
}
