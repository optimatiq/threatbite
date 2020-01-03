package email

import (
	"regexp"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/email/datasource"
)

type free struct {
	domain *datasource.Domain
}

func newFree(source datasource.DataSource) *free {
	return &free{domain: datasource.NewDomain(source, "disposal")}
}

var reFreeSubDomains = regexp.MustCompile(".hub.pl$|.int.pl$")

func (d *free) isFree(email string) bool {
	domain := strings.ToLower(strings.Split(email, "@")[1])
	isfree := d.domain.Check(domain) || reFreeSubDomains.MatchString(domain)
	log.Debugf("[isFree] domain: %s free: %t", domain, isfree)
	return isfree
}
