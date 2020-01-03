package controllers

import (
	"regexp"
	"strings"

	lru "github.com/hashicorp/golang-lru"
	"github.com/optimatiq/threatbite/email"
)

// EmailResult response object, which contains detailed information returned from Check method.
type EmailResult struct {
	Scoring       uint8 `json:"scoring"`
	AccountExists bool  `json:"exists"`
	CatchAll      bool  `json:"catchall"`
	DefaultUser   bool  `json:"default"`
	Disposal      bool  `json:"disposal"`
	Free          bool  `json:"free"`
	Leaked        bool  `json:"leaked"`
	Valid         bool  `json:"valid"`
}

// Email is a controller container with the cache.
type Email struct {
	cache     *lru.Cache
	emailInfo *email.Email
}

// NewEmail creates an Email scoring controller.
func NewEmail(emailInfo *email.Email) (*Email, error) {
	cache, err := lru.New(4096)
	if err != nil {
		return nil, err
	}

	return &Email{
		cache:     cache,
		emailInfo: emailInfo,
	}, nil
}

var reEmail = regexp.MustCompile("(?i)^[a-z0-9!#$%&'*+/i=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$")

// Validate returns nil if provided email is valid,
// otherwise, returns error, which can be presented to the user.
func (e *Email) Validate(email string) error {
	if !reEmail.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// Check is the main module functions, which is used to perform all checks for given argument.
func (e *Email) Check(address string) (*EmailResult, error) {
	if len(strings.Split(address, "@")) != 2 {
		return nil, ErrInvalidEmail
	}

	if v, ok := e.cache.Get(address); ok {
		return v.(*EmailResult), nil
	}

	info := e.emailInfo.GetInfo(address)

	result := &EmailResult{
		Scoring:       info.EmailScoring,
		AccountExists: info.IsExistingAccount,
		CatchAll:      info.IsCatchAll,
		DefaultUser:   info.IsDefaultUser,
		Disposal:      info.IsDisposal,
		Free:          info.IsFree,
		Leaked:        info.IsLeaked,
		Valid:         info.IsValid,
	}

	if !e.cache.Contains(address) {
		e.cache.Add(address, result)
	}

	return result, nil
}
