package datasource

// ListDataSource stores current state (counters) of this source.
type ListDataSource struct {
	i    int
	list []string
}

// NewListDataSource first argument is a list of domains.
// Returns DataSource or error on IP parsing.
func NewListDataSource(list []string) *ListDataSource {
	return &ListDataSource{
		list: list,
	}
}

// Reset rewinds source to the beginning.
func (s *ListDataSource) Reset() error {
	s.i = 0
	return nil
}

// Next returns domain from the provided list in NewListDataSource method.
// ErrNoData is returned when there is no data, this error indicates that we reached the end.
func (s *ListDataSource) Next() (string, error) {
	if s.i >= len(s.list) || len(s.list) <= 0 {
		return "", ErrNoData
	}

	v := s.list[s.i]
	s.i++
	return v, nil
}
