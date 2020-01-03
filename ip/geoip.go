package ip

import "net"

type geoip interface {
	getCountry(ip net.IP) (string, error)
	getCompany(ip net.IP) (string, error)
	update() error
}
