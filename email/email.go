package email

import (
	"context"
	"crypto/md5" // #nosec
	"encoding/hex"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/optimatiq/threatbite/email/datasource"

	isd "github.com/jbenet/go-is-domain"
	"github.com/labstack/gommon/log"
	"golang.org/x/sync/errgroup"
)

const pwnedAPI = "https://haveibeenpwned.com/api/v3/breachedaccount/"

// Info a struct, which contains information about email address.
type Info struct {
	EmailScoring      uint8
	IsDisposal        bool
	IsDefaultUser     bool
	IsFree            bool
	IsValid           bool
	IsCatchAll        bool
	IsExistingAccount bool
	IsLeaked          bool
}

// Email container for email service.
type Email struct {
	pwnedKey  string
	smtpHello string
	smtpFrom  string
	disposal  *disposal
	free      *free
}

// NewEmail returns email service, which is used to get detailed information about email address.
func NewEmail(pwnedKey, smtpHello, smtpFrom string, disposalSources, freeSources datasource.DataSource) *Email {
	if pwnedKey == "" {
		log.Infof("[email] Haveibeenpwned license is not present, reputation accuracy is degraded.")
	}

	if smtpFrom == "" || smtpHello == "" {
		log.Infof("[email] SMTP configuration is not present, reputation accuracy is degraded.")
	}

	return &Email{
		pwnedKey:  pwnedKey,
		smtpFrom:  smtpFrom,
		smtpHello: smtpHello,
		disposal:  newDisposal(disposalSources),
		free:      newFree(freeSources),
	}
}

// GetInfo returns computed information (Info struct) for given email address.
func (e *Email) GetInfo(email string) Info {
	var g errgroup.Group

	var isDisposal bool
	g.Go(func() (err error) {
		isDisposal = e.disposal.isDisposal(email)
		return
	})

	var isUserDefault bool
	g.Go(func() (err error) {
		isUserDefault = e.isUserDefault(email)
		return
	})

	var isFree bool
	g.Go(func() (err error) {
		isFree = e.free.isFree(email)
		return
	})

	// isRFC used in isValid
	var isRFC bool
	g.Go(func() (err error) {
		isRFC = e.isRFC(email)
		return
	})

	// isDomainIANA used in isValid
	var isDomainIANA bool
	g.Go(func() (err error) {
		isDomainIANA = e.isDomainIANA(email)
		return
	})

	var isCatchAll bool
	g.Go(func() (err error) {
		isCatchAll = e.isCatchAll(email)
		return
	})

	var isExisting bool
	g.Go(func() (err error) {
		isExisting = e.isExisting(email)
		return
	})

	var isPwned bool
	g.Go(func() (err error) {
		isPwned = e.isPwned(email)
		return
	})

	_ = g.Wait() // none of the goroutines return error, so we don't need to check it.

	var isValid bool
	if isRFC && isDomainIANA {
		isValid = true
	}

	// Calculate scoring 0-100 (worst-best)
	var scoring uint8 = 80

	// Free email accounts have reduced trust
	if isFree {
		scoring -= 10
	}
	if !isFree {
		scoring += 10
	}

	// Default users have very low trust
	if isUserDefault {
		scoring -= 35
	}
	if !isUserDefault {
		scoring += 3
	}

	if isDisposal {
		scoring -= 45
	}
	if !isDisposal {
		scoring += 4
	}

	if isCatchAll {
		scoring -= 30
	}
	if !isCatchAll {
		scoring += 8
	}

	// If email has leaked it means that it has history and exists
	if isPwned {
		scoring += 3
	}
	if !isPwned {
		scoring--
	}

	if isDomainIANA {
		scoring++
	}
	if !isDomainIANA {
		scoring = 0
	}

	if isExisting {
		scoring += 2
	}
	if !isExisting {
		scoring = 0
	}

	// If email is not proper set to zero
	if isRFC {
		scoring++
	}
	if !isRFC {
		scoring = 0
	}

	// Maximum value is 100
	if scoring > 100 {
		scoring = 100
	}

	return Info{
		EmailScoring:      scoring,
		IsDisposal:        isDisposal,
		IsDefaultUser:     isUserDefault,
		IsFree:            isFree,
		IsValid:           isValid,
		IsCatchAll:        isCatchAll,
		IsExistingAccount: isExisting,
		IsLeaked:          isPwned,
	}
}

// isRFC checks if the length of "local part" (before the "@") which maximum is 64 characters (octets)
// and the length of domain part (after the "@") which maximum is 255 characters (octets)
func (e *Email) isRFC(email string) bool {
	valid := len(strings.Split(email, "@")[0]) <= 64 && len(strings.Split(email, "@")[1]) <= 255
	log.Debugf("[isRFC] email: %s - %t", email, valid)
	return valid
}

// isDomainIANA checks if TLD is on IANA http://data.iana.org/TLD/tlds-alpha-by-domain.txt
func (e *Email) isDomainIANA(email string) bool {
	valid := isd.IsDomain(strings.Split(email, "@")[1])
	log.Debugf("[isDomainIANA] email: %s - %t", email, valid)
	return valid
}

// isUserDefault checks is user Name is default, suspicious or commonly used as spamtrap
func (e *Email) isUserDefault(email string) bool {
	user := strings.Split(email, "@")[0]
	_, found := defaultUsernames[strings.ToLower(user)]
	log.Debugf("[isUserDefault] email: %s - %t", email, found)
	return found
}

// checkDomainMX checks if domain have configured MX record and returns IP with the highest priority
func (e *Email) getDomainMX(email string) (string, error) {
	mxRecords, err := lookupMXWithTimeout(strings.Split(email, "@")[1], 1*time.Second)
	log.Debugf("[checkDomainMX] email: %s mxRecords: %s, error: %s", email, mxRecords, err)
	if err != nil {
		return "", err
	}
	return mxRecords[0].Host, nil
}

// getDomainIP checks if domain has at least one A record and return it
func (e *Email) getDomainIP(email string) (string, error) {
	records, err := lookupIPWithTimeout(strings.Split(email, "@")[1], 1*time.Second)
	log.Debugf("[getDomainIP] email: %s records: %v, error: %s", email, records, err)
	if err != nil {
		return "", err
	}

	return records[0].String(), nil
}

// getRandomUser generates random user name
func (e *Email) getRandomUser() string {
	now := time.Now()
	h := md5.New() // #nosec
	_, err := h.Write([]byte(now.String()))
	if err != nil {
		return now.Format("01-02-2006")
	}
	return hex.EncodeToString(h.Sum(nil))
}

// isCatchAll checks if remote server is configured as Catch all
func (e *Email) isCatchAll(email string) bool {
	domain := strings.ToLower(strings.Split(email, "@")[1])
	return e.isExisting(e.getRandomUser() + "@" + domain)
}

var reSMTP4xx = regexp.MustCompile("^4")
var reSMTP5xx = regexp.MustCompile("^5")

// isExisting checks if account exists on remote server
func (e *Email) isExisting(email string) bool {
	if e.smtpHello == "" || e.smtpFrom == "" {
		log.Debug("[isExisting] SMTP is not configured, not checking")
		return false
	}

	/*
		We are working as MTA (server) so by default we are connecting to 25/tcp without TLS
		SMTP session example:

		$ telnet example.org 25
		S: 220 example.org ESMTP Sendmail 8.13.1/8.13.1; Wed, 30 Aug 2006 07:36:42 -0400
		C: HELO mailout1.phrednet.com
		S: 250 example.org Hello ip068.subnet71.gci-net.com [216.183.71.68], pleased to meet you
		C: MAIL FROM:<xxxx@example.com>
		S: 250 2.1.0 <xxxx@example.com>... Sender ok
		C: RCPT TO:<yyyy@example.com>
		S: 250 2.1.5 <yyyy@example.com>... Recipient ok
		C: DATA
		S: 354 Enter mail, end with "." on a line by itself
		From: Dave\r\nTo: David\r\nSubject: Its not SPAM\r\n\r\nThis is message 1.\r\n.\r\n
		S: 250 2.0.0 k7TKIBYb024731 Message accepted for delivery
		C: QUIT
		S: 221 2.0.0 example.org closing connection
		Connection closed by foreign host.
		$

		In this implementation we are sending only 'MAIL FROM' and 'RCPT TO' to get 250 response. In this case recipient
		will not receive any notification because we are not sending 'DATA' command.

		Proper SMTP session ends with 'QUIT' command. To avoid logging on the remote server, we close the connection
		without sending the 'QUIT' command.

		- First 'HELO' command must have FQDN argument. Many servers are checking (to block spammers) if this name exists
		and is directed to the IP address from which the connection comes.

		- Email used in 'MAIL FROM' must have a domain that exists. Many anti-spam systems also check SPF and DKIM if they
		are configured in DNS.

		- IP address from which the connection comes must have configured PTR (revDNS) and domain.

		- RevDNS shouldn't have strings like `vps`, `cloud`, `virtual` or `static` in the name. Such connection can be
		blocked by the remote server (like mail.ru).

		- Before starting using IP check if it's not listed on remote RBL
	*/

	lowerEmail := strings.ToLower(email)

	var connHost string

	connMX, errMX := e.getDomainMX(lowerEmail)
	if errMX != nil {
		connIP, errIP := e.getDomainIP(lowerEmail)
		if errIP != nil {
			return false
		}
		if connIP != "" {
			connHost = connIP
		}
	}
	if connMX != "" {
		connHost = connMX
	}

	// TODO(PG) Check 465, 587 and STARTTLS
	connDial, err := net.DialTimeout("tcp", connHost+":25", time.Duration(3)*time.Second)
	if err != nil {
		log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		return false
	}

	connSMTP, err := smtp.NewClient(connDial, connHost+":25")
	if err != nil {
		log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		return false
	}

	if err := connSMTP.Hello(e.smtpHello); err != nil {
		log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		return false
	}

	if err := connSMTP.Mail(e.smtpFrom); err != nil {
		log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		return false
	}

	resp := connSMTP.Rcpt(lowerEmail)
	if resp != nil {
		if reSMTP5xx.MatchString(resp.Error()) {
			log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
			return false
		}

		if reSMTP4xx.MatchString(resp.Error()) {
			// We can make another test after 1 min to bypass Grey Listing
			// But we need to implement tokens and repeat this test from the same IP
			log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		}
	}

	err = connSMTP.Close()
	if err != nil {
		log.Debugf("[isExisting] email: %s host: %v, error: %s", email, connHost, err)
		return false
	}

	return true
}

func (e *Email) isPwned(email string) bool {
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}

	request, err := http.NewRequest("GET", pwnedAPI+email, nil)
	if err != nil {
		log.Debugf("[isPwned] cannot prepare request, error: %s", err)
		return false
	}
	request.Header.Set("hibp-api-key", e.pwnedKey)

	response, err := netClient.Do(request)
	if err != nil {
		log.Errorf("[isPwned] cannot make request, error: %s", err)
		return false
	}

	if response.StatusCode != http.StatusOK {
		log.Errorf("[isPwned] invalid status code %d", response.StatusCode)
		return false
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debugf("[isPwned] cannot read response, error: %s", err)
		return false
	}

	if len(body) > 0 {
		log.Debugf("[isPwned] email: %s - %s", email, body)
		return true
	}

	return false
}

// RunUpdates schedules and runs updates.
// Update interval is defined for each source individually.
func (e *Email) RunUpdates() {
	ctx := context.Background()

	runAndSchedule(ctx, 24*time.Hour, func() {
		if err := e.disposal.domain.Load(); err != nil {
			log.Error(err)
		}
	})

	runAndSchedule(ctx, 24*time.Hour, func() {
		if err := e.free.domain.Load(); err != nil {
			log.Error(err)
		}
	})
}

func runAndSchedule(ctx context.Context, interval time.Duration, f func()) {
	go func() {
		t := time.NewTimer(0) // first run - immediately
		for {
			select {
			case <-t.C:
				f()
				t = time.NewTimer(interval) // next runs according to the schedule
			case <-ctx.Done():
				return
			}
		}
	}()
}
