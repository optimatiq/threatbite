package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkPrivateIP(t *testing.T) {
	tests := map[string]bool{
		"1.1.1.1":         false,
		"255.255.255.255": true,
		"8.8.8.8":         false,
		"127.0.0.1":       true,
		"192.168.10.10":   true,
		"10.0.0.1":        true,
		"123.123.123.12":  false,
	}

	for ip, v := range tests {
		assert.Equal(t, isPrivateIP(net.ParseIP(ip)), v, ip)
	}
}
