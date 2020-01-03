package datasource

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// DirectoryDataSource stores current state (counters, files, scanners) of this source.
type DirectoryDataSource struct {
	path    string
	files   []string
	f       int
	fLock   sync.Mutex
	scanner *bufio.Scanner
	file    *os.File
}

// NewDirectoryDataSource returns iterator, which looks for all *.txt files in given directory or error.
// Files should have each IPv4 IPv6 or CIDR in new line.
// Comments are allowed and ignored. Comments start with # at the beginning of the line.
func NewDirectoryDataSource(directory string) (*DirectoryDataSource, error) {
	dataSource := &DirectoryDataSource{
		path: path.Join(directory, "*.txt"),
	}

	err := dataSource.loadFiles()
	if err != nil {
		return nil, err
	}

	return dataSource, nil
}

// Reset rewinds source to the beginning.
func (s *DirectoryDataSource) Reset() error {
	return s.loadFiles()
}

func (s *DirectoryDataSource) loadFiles() error {
	s.fLock.Lock()
	defer s.fLock.Unlock()

	matches, err := filepath.Glob(s.path)
	if err != nil {
		return fmt.Errorf("directory: %s, error: %w", s.path, err)
	}

	if len(matches) <= 0 {
		return fmt.Errorf("no data in the directory: %s, error: %w", s.path, ErrInvalidData)
	}

	s.scanner = nil
	s.files = matches
	s.f = 0

	return nil
}

// Next returns IP/CIDR, this method knows, which file and line needs to be read.
// ErrNoData is returned when there is no data, this error indicates that we reached the end.
func (s *DirectoryDataSource) Next() (*net.IPNet, error) {
	if s.f >= len(s.files) || len(s.files) <= 0 {
		return nil, ErrNoData
	}

	// #nosec G304
	if s.scanner == nil {
		filename := s.files[s.f]
		file, err := os.Open(filename)
		if err != nil {
			s.f++
			return nil, fmt.Errorf("filename: %s, error: %w", filename, err)
		}
		s.scanner = bufio.NewScanner(file)
		s.file = file
	}

	var line string
	for s.scanner.Scan() {
		line = s.scanner.Text()

		// Comment
		if strings.Index(line, "#") == 0 {
			continue
		}

		// CIDR
		if strings.Contains(line, "/") {
			_, ipNet, err := net.ParseCIDR(line)
			if err != nil {
				return nil, ErrInvalidData
			}
			return ipNet, nil
		}

		// Single IP
		ip := net.ParseIP(line)
		if ip == nil {
			return nil, ErrInvalidData
		}
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(8*len(ip), 8*len(ip))}, nil
	}

	// End of file or error
	_ = s.file.Close()
	err := s.scanner.Err()
	s.scanner = nil

	if err != nil {
		return nil, err
	}

	s.f++
	return s.Next()
}
