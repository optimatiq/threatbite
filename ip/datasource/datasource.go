package datasource

import (
	"errors"
	"net"
)

// ErrNoData no more date in iterator, means that we finished iterating.
var ErrNoData = errors.New("no more data")

// ErrInvalidData source is available or data provided in the source were not valid IPv4, IPv6 or CIDR.
// When this error is return Next() method is called again.
var ErrInvalidData = errors.New("invalid data")

// DataSource defines method for accessing stream of addresses.
type DataSource interface {
	// Next returns net.IPNet on success or error.
	// ErrNoData and ErrInvalidData can be ignored
	Next() (*net.IPNet, error)
	Reset() error
}
