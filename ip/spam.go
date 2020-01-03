package ip

import (
	"net"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/ip/datasource"
)

type spam struct {
	ipnet *datasource.IPNet
}

func newSpam(source datasource.DataSource) *spam {
	return &spam{ipnet: datasource.NewIPNet(source, "spam")}
}

func (s *spam) isSpam(ip net.IP) (bool, error) {
	isSpam, err := s.ipnet.Check(ip)
	log.Debugf("[isSpam] ip: %s tor: %t", ip, isSpam)
	return isSpam, err
}
