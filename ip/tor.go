package ip

import (
	"fmt"
	"io/ioutil"
	"net"
	"regexp"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/ip/datasource"
)

// TODO(RW) we should use different list of exit nodes, official endpoint can contain outdated data.
const torExitNodes = "https://check.torproject.org/exit-addresses"

type tor struct {
	ipnet *datasource.IPNet
}

func newTor() *tor {
	return &tor{ipnet: datasource.NewIPNet(datasource.NewEmptyDataSource(), "tor")}
}

func (t *tor) update() error {
	log.Debug("[tor] update start")
	defer log.Debug("[tor] update finished")

	response, err := defaultHTTPClient.Get(torExitNodes)
	if err != nil {
		return fmt.Errorf("cannot download TOR exit nodes, error: %w", err)
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("cannot read body of TOR exit nodes, error: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("cannot close response body, error: %w", err)
	}

	reExitNode := regexp.MustCompile(`ExitAddress (\d+\.\d+\.\d+\.\d+)`)
	var nodes []string
	for _, node := range reExitNode.FindAllStringSubmatch(string(content), -1) {
		nodes = append(nodes, node[1])
	}

	ds, err := datasource.NewListDataSource(nodes)
	if err != nil {
		return fmt.Errorf("cannot create datasource for TOR, error: %w", err)
	}

	t.ipnet = datasource.NewIPNet(ds, "tor")
	return t.ipnet.Load()
}

func (t *tor) isTor(ip net.IP) (bool, error) {
	isTor, err := t.ipnet.Check(ip)
	log.Debugf("[checkTor] ip: %s tor: %t", ip, isTor)
	return isTor, err
}
