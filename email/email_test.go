package email

import (
	"testing"

	"github.com/optimatiq/threatbite/email/datasource"

	"github.com/stretchr/testify/assert"
)

func Test_checkLocalLength(t *testing.T) {
	tests := map[string]bool{
		"mail@example.com":      true,
		"m_il.MaIl@exam_le.com": true,
		"63.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.@ExapLe.CoM":    true,
		"64.M.aI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.@ExapLe.CoM":   true,
		"65.M.aI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.L@ExapLe.CoM":  false,
		"66.M.aI.M.aI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.MaI.L@ExapLe.CoM": false,
	}

	e := NewEmail("", "D", "", nil, nil)
	for mail, v := range tests {
		assert.Equal(t, v, e.isRFC(mail), mail)
	}
}

func Test_checkDomainLength(t *testing.T) {
	tests := map[string]bool{
		"mail@example.com":      true,
		"m_il.MaIl@exam-le.com": true,
		"Mail@255e.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.N.com":  true,
		"Mail@256e.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Name.Na.com": false,
	}

	e := NewEmail("", "D", "", nil, nil)
	for mail, v := range tests {
		assert.Equal(t, v, e.isRFC(mail), mail)
	}

	assert.Panics(t, func() { e.isRFC("invalid_mail") })
}

func Test_checkDomainIANA(t *testing.T) {
	tests := map[string]bool{
		"mail@example.com":      true,
		"m_il.MaIl@exam-le.com": true,
		"Mail@example.Com":      true,
		"Mail@example.Pizza1":   false,
		"Mail@0.0":              false,
	}

	e := NewEmail("", "D", "", nil, nil)
	for mail, v := range tests {
		assert.Equal(t, v, e.isDomainIANA(mail), mail)
	}

	assert.Panics(t, func() { e.isDomainIANA("invalid_mail") })
}

func Test_checkUserDefault(t *testing.T) {
	tests := map[string]bool{
		"mail@example.com":      true,
		"anti.spam@exam-le.com": false,
		"default@example.Com":   true,
		"DeFAult@example.com":   true,
		"AntiSpam@example.com":  true,
	}

	e := NewEmail("", "D", "", nil, nil)
	for mail, v := range tests {
		assert.Equal(t, e.isUserDefault(mail), v, mail)
	}
}

func Test_checkDisposal(t *testing.T) {
	tests := map[string]bool{
		"foo@0-mail.com":                   true,
		"anti.spam@example-reputation.com": false,
		"default@niepodam.pl":              true,
		"DeFAult@NiEpOdam.PL":              true,
		"AntiSpam@126.COM":                 true,
	}

	e := NewEmail("", "", "", datasource.NewListDataSource([]string{"0-mail.com", "niepodam.pl", "126.com"}), datasource.NewEmptyDataSource())
	err := e.disposal.domain.Load()
	assert.NoError(t, err)

	for mail, v := range tests {
		assert.Equal(t, v, e.disposal.isDisposal(mail), mail)
	}
}

func Test_checkFree(t *testing.T) {
	tests := map[string]bool{
		"foo@wp.pl":                        true,
		"anti.spam@example-reputation.com": false,
		"default@gmail.com":                true,
		"DeFAult@GmAil.Com":                true,
		"AntiSpam@YAHOO.COM":               true,
	}
	e := NewEmail("", "", "", datasource.NewEmptyDataSource(), datasource.NewListDataSource([]string{"wp.pl", "gmail.com", "YAHOO.COM"}))
	err := e.free.domain.Load()
	assert.NoError(t, err)

	for mail, v := range tests {
		assert.Equal(t, v, e.free.isFree(mail), mail)
	}
}

func Test_isPwned(t *testing.T) {
	tests := map[string]bool{
		"default@gmail.com": false,
	}
	e := NewEmail("invalid_key", "D", "", nil, nil)
	for mail, v := range tests {
		assert.Equal(t, v, e.isPwned(mail), mail)
	}
}
