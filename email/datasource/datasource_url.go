package datasource

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

// URLDataSource stores current state (counters, URLs, scanners) of this source.
type URLDataSource struct {
	urls    []string
	u       int
	scanner *bufio.Scanner
	client  *http.Client
}

// NewURLDataSource returns iterator, which downloads lists from provided URLs and extract addresses.
// Comments are allowed and ignored. Comments start with # at the beginning of the line.
// Some lists have comments after their address, they are also ignored
func NewURLDataSource(urls []string) *URLDataSource {
	dataSource := &URLDataSource{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   60 * time.Second,
					KeepAlive: 15 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   60 * time.Second,
				ExpectContinueTimeout: 10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
			},
			Timeout: 120 * time.Second,
		},
		urls: urls,
	}

	return dataSource
}

// Reset rewinds source to the beginning.
func (s *URLDataSource) Reset() error {
	s.u = 0
	s.scanner = nil
	return nil
}

func (s *URLDataSource) Next() (string, error) {
	if s.u >= len(s.urls) || len(s.urls) <= 0 {
		return "", ErrNoData
	}
	url := s.urls[s.u]

	if s.scanner == nil {
		response, err := s.client.Get(url)
		if err != nil {
			log.Errorf("[datasource] cannot download list from: %s, error: %s", url, err)
			s.u++
			return "", ErrInvalidData
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Errorf("[datasource] cannot read from: %s, error: %s", url, err)
			s.u++
			return "", ErrInvalidData
		}

		if err := response.Body.Close(); err != nil {
			log.Errorf("[datasource] cannot close response body from: %s, error: %s", url, err)
		}

		s.scanner = bufio.NewScanner(bytes.NewReader(body))
	}

	var line string
	for s.scanner.Scan() {
		line = s.scanner.Text()
		// some lists have address with optional comment as a second argument separated by spaces or tabs
		line = strings.ReplaceAll(line, "\t", " ")
		line := strings.Split(line, " ")[0]

		// Comment
		if strings.Index(line, "#") == 0 {
			continue
		}

		return line, nil
	}

	err := s.scanner.Err()
	s.scanner = nil

	if err != nil {
		return "", err
	}

	s.u++
	return s.Next()
}
