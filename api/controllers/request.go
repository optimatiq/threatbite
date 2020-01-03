package controllers

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"encoding/json"
	"net"

	"github.com/go-playground/validator"
	lru "github.com/hashicorp/golang-lru"
	"github.com/optimatiq/threatbite/browser"
	"github.com/optimatiq/threatbite/ip"
)

// RequestResult response object, which contains detailed information returned from Check method.
type RequestResult struct {
	IPResult
	browser.UserAgent
	Bot    bool
	Mobile bool
	Script bool
}

// RequestQuery struct, which is used to calculate scoring for given request (based on HTTP values).
// Some fields are required (IP, Host, URI, Method, UserAgent) other are options.
type RequestQuery struct {
	// Required fields
	IP        string `json:"ip" form:"ip" validate:"required,ip"`
	Host      string `json:"host" form:"host" validate:"required,hostname"`
	URI       string `json:"uri" form:"uri" validate:"required,uri"`
	Method    string `json:"method" form:"method" validate:"required,oneof=GET HEAD POST PUT DELETE TRACE OPTIONS PATCH"`
	UserAgent string `json:"user_agent" form:"user_agent" validate:"required"`

	// Optional fields
	Protocol    string            `json:"protocol" form:"protocol"`
	Scheme      string            `json:"scheme" form:"scheme" validate:"omitempty,oneof=http https"`
	ContentType string            `json:"content_type" form:"content_type"`
	Headers     map[string]string `json:"headers" form:"headers"`
}

func (r RequestQuery) hash() (string, error) {
	key, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	h := md5.New() // #nosec
	_, err = h.Write(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Request is a container for HTTP request controller.
type Request struct {
	ipinfo    *ip.IP
	cache     *lru.Cache
	validator *validator.Validate
}

// NewRequest creates new HTTP request scoring module.
func NewRequest(ipinfo *ip.IP) (*Request, error) {
	cache, err := lru.New(4096)
	if err != nil {
		return nil, err
	}

	request := &Request{
		validator: validator.New(),
		cache:     cache,
		ipinfo:    ipinfo,
	}

	return request, nil
}

// Validate returns nil if provided data is valid,
// otherwise, returns error, which can be presented to the user.
// Validation rules are defined as a struct tags.
func (r *Request) Validate(request RequestQuery) error {
	// TODO translate errors, now they are connected with RequestQuery struct fields as internal representation
	//  - example: 'RequestQuery.IP' Error:Field validation for 'IP' failed on the 'required' tag
	return r.validator.Struct(request)
}

// Check is the main module functions which is used to perform all checks for given argument.
func (r *Request) Check(request RequestQuery) (*RequestResult, error) {
	// TODO add business logic
	key, err := request.hash()
	if err != nil {
		return nil, err
	}

	if v, ok := r.cache.Get(key); ok {
		return v.(*RequestResult), nil
	}

	ip := net.ParseIP(request.IP)
	if ip == nil {
		return nil, ErrInvalidIP
	}

	info, err := r.ipinfo.GetInfo(ip)
	if err != nil {
		return nil, err
	}

	result := &RequestResult{
		IPResult: IPResult{
			Country:      info.Country,
			Tor:          info.IsTor,
			Proxy:        info.IsProxy,
			SearchEngine: info.IsSearchEngine,
			Private:      info.IsPrivate,
			Spam:         info.IsSpam,
			Datacenter:   info.IsDatacenter,
		},
		UserAgent: *browser.GetUserAgent(request.UserAgent),
		Bot:       browser.IsBotUserAgent(request.UserAgent),
		Mobile:    browser.IsMobileUserAgent(request.UserAgent),
		Script:    browser.IsScriptUserAgent(request.UserAgent),
	}
	if !r.cache.Contains(key) {
		r.cache.Add(key, result)
	}
	return result, nil
}
