package datasource

import "net"

// EmptyDataSource as the name suggest, this data source contains no data.
type EmptyDataSource struct {
}

// NewEmptyDataSource returns empty data source
func NewEmptyDataSource() *EmptyDataSource {
	return &EmptyDataSource{}
}

// Reset does nothing.
func (s *EmptyDataSource) Reset() error {
	return nil
}

// Next returns ErrNoData error always
func (s *EmptyDataSource) Next() (*net.IPNet, error) {
	return nil, ErrNoData
}
