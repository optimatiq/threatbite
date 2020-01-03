package datasource

import (
	"fmt"
	"strings"

	"github.com/labstack/gommon/log"
)

type Domain struct {
	domains map[string]bool
	ds      DataSource
	name    string
}

// NewDomain returns a new domain list build on top of map.
// Load method has to be called manually in order to get data from data source
func NewDomain(ds DataSource, name string) *Domain {
	return &Domain{
		domains: make(map[string]bool),
		ds:      ds,
		name:    name,
	}
}

// Check if lists contains domain from request
func (d *Domain) Check(domain string) bool {
	_, found := d.domains[domain]
	return found
}

func (d *Domain) Load() error {
	log.Debugf("[list] loading %s list start", d.name)
	defer func() {
		log.Debugf("[list] loading %s stop; stats domains: %d", d.name, len(d.domains))
	}()

	if err := d.ds.Reset(); err != nil {
		return fmt.Errorf("could not reset data source, error: %w", err)
	}

	for {
		domain, err := d.ds.Next()
		if err != nil {
			if err == ErrNoData {
				return nil
			} else if err == ErrInvalidData {
				continue
			} else {
				return fmt.Errorf("could not iterate over data source, error: %w", err)
			}
		}

		d.domains[strings.ToLower(domain)] = true
	}
}
