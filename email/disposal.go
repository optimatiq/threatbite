package email

import (
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/email/datasource"
)

type disposal struct {
	domain *datasource.Domain
}

func newDisposal(source datasource.DataSource) *disposal {
	return &disposal{domain: datasource.NewDomain(source, "disposal")}
}

func (d *disposal) isDisposal(email string) bool {
	domain := strings.ToLower(strings.Split(email, "@")[1])
	isDisposal := d.domain.Check(domain)
	log.Debugf("[isDisposal] domain: %s disposal: %t", domain, isDisposal)
	return isDisposal
}
