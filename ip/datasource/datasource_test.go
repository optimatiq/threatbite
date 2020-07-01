package datasource

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DatasourceSuite struct {
	suite.Suite
	privateRand *rand.Rand
}

func (suite *DatasourceSuite) SetupTest() {
	suite.privateRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (suite *DatasourceSuite) Test_StringDatasource() {
	tests := []struct {
		list    []string
		wantErr bool
	}{
		{
			list:    []string{""},
			wantErr: true,
		},
		{
			list:    []string{"a.b.c.d"},
			wantErr: true,
		},
		{
			list:    []string{"127.0.0.1"},
			wantErr: false,
		},
		{
			list:    []string{"127.0.0.1/8"},
			wantErr: false,
		},
		{
			list:    []string{"127.0.0.1/8", "1.2.3.4"},
			wantErr: false,
		},
		{
			list:    []string{"127.0.0.1/8", "1.2.3.4/32"},
			wantErr: false,
		},
		{
			list:    []string{"127.0.0.1/8", "1.2.3.4/32", "invalid"},
			wantErr: true,
		},
		{
			list:    []string{"127.0.0.1/8", "1.2.3.4/32", "invalid_cidr/12"},
			wantErr: true,
		},
	}
	for _, t := range tests {
		_, err := NewListDataSource(t.list)
		if t.wantErr {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}
	}
}

func (suite *DatasourceSuite) Test_StringDatasourceFunctions() {
	ds, err := NewListDataSource([]string{"127.0.0.1/8", "1.2.3.4/32"})
	suite.NoError(err)
	ip, err := ds.Next()
	suite.NoError(err)
	_, expected1, err := net.ParseCIDR("127.0.0.1/8")
	suite.NoError(err)
	suite.Equal(expected1, ip)

	ip, err = ds.Next()
	suite.NoError(err)
	_, expected2, err := net.ParseCIDR("1.2.3.4/32")
	suite.NoError(err)
	suite.Equal(expected2, ip)

	ip, err = ds.Next()
	suite.Error(err)
	suite.Nil(ip)

	err = ds.Reset()
	suite.NoError(err)

	ip, err = ds.Next()
	suite.NoError(err)
	suite.Equal(expected1, ip)
}

func (suite *DatasourceSuite) Test_NewDirectoryDatasource() {
	dir := suite.createDir(true)
	_, err := NewDirectoryDataSource(dir)
	suite.NoError(err)
	err = os.RemoveAll(dir)
	suite.NoError(err)

	dir = suite.createDir(false)
	_, err = NewDirectoryDataSource(dir)
	suite.Error(err)
	err = os.RemoveAll(dir)
	suite.NoError(err)
}

func (suite *DatasourceSuite) Test_NewURLDataSource() {
	ds := NewURLDataSource([]string{
		"https://iplists.firehol.org/files/proxz_1d.ipset",
	})

	ip, err := ds.Next()
	suite.NoError(err)
	suite.NotEmpty(ip)

	ds = NewURLDataSource([]string{
		"invalid",
	})

	suite.NoError(err)
	ip, err = ds.Next()
	suite.Error(err)
	suite.Empty(ip)
}

func (suite *DatasourceSuite) Test_DirectoryDatasourceNext() {
	dir := suite.createDir(true)
	ds, err := NewDirectoryDataSource(dir)
	suite.NoError(err)
	for {
		ip, err := ds.Next()
		if err == ErrNoData {
			break
		} else if err == ErrInvalidData {
			continue
		}
		suite.NoError(err)
		suite.NotEmpty(ip)
	}
	err = ds.Reset()
	suite.NoError(err)

	for {
		ip, err := ds.Next()
		if err == ErrNoData {
			break
		} else if err == ErrInvalidData {
			continue
		}
		suite.NoError(err)
		suite.NotEmpty(ip)
	}

	err = os.RemoveAll(dir)
	suite.NoError(err)

	_, err = NewDirectoryDataSource("\\")
	suite.Error(err)
}

func (suite *DatasourceSuite) createDir(content bool) string {
	dir, err := ioutil.TempDir("", "datasource")
	suite.NoError(err)
	if content {
		for _, name := range []string{"1", "2", "3"} {
			tmpfn := filepath.Join(dir, name+".txt")
			b := bytes.NewBuffer(nil)
			for i := 0; i < 100; i++ {
				b.Write([]byte("# comment\n"))
				b.Write([]byte("invalid data x.x.x.x\n"))
				b.Write([]byte(suite.randomIPv4(false) + "\n"))
				b.Write([]byte(suite.randomIPv4(true) + "\n"))
				b.Write([]byte(suite.randomIPv6() + "\n"))
			}
			err = ioutil.WriteFile(tmpfn, b.Bytes(), 0600)
			suite.NoError(err)
		}
	}

	return dir
}

func (suite *DatasourceSuite) randomIPv4(cidr bool) string {
	var blocks []string

	for i := 0; i < net.IPv4len; i++ {
		number := suite.privateRand.Intn(255)
		blocks = append(blocks, strconv.Itoa(number))
	}

	ip := strings.Join(blocks, ".")
	if cidr {
		return fmt.Sprintf("%s/%d", ip, suite.privateRand.Intn(32))
	}
	return ip
}

func (suite *DatasourceSuite) randomIPv6() string {
	var ip net.IP
	for i := 0; i < net.IPv6len; i++ {
		number := uint8(suite.privateRand.Intn(255))
		ip = append(ip, number)
	}
	return ip.String()
}

func TestDatasourceSuite(t *testing.T) {
	suite.Run(t, new(DatasourceSuite))
}
