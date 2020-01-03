package datasource

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListSuite struct {
	suite.Suite
}

func (suite *ListSuite) Test_CheckMixed() {
	tests := []struct {
		data    []string
		check   string
		want    bool
		wantErr bool
	}{

		{
			[]string{"2001:db8:1234::/48", "1.1.1.1/8"},
			"2001:db8:1234:0:0:8a2e:370:7334",
			true,
			false,
		},
		{
			[]string{"2001:db8:1234::/48", "1.1.1.1/8"},
			"2001:db8:1234:0:0:8a2e:370:7334",
			true,
			false,
		},
	}
	for _, t := range tests {
		ds, err := NewListDataSource(t.data)
		suite.NoError(err)

		ipnet := NewIPNet(ds, "testList")
		err = ipnet.Load()
		suite.NoError(err)

		v, err := ipnet.Check(net.ParseIP(t.check))

		if t.wantErr {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}

		suite.Equal(t.want, v, t)
	}
}

func (suite *ListSuite) Test_CheckCIDR() {
	tests := []struct {
		data    []string
		check   string
		want    bool
		wantErr bool
	}{

		{
			[]string{"2001:db8:1234::/48"},
			"2001:db8:1234:0:0:8a2e:370:7334",
			true,
			false,
		},

		{
			[]string{"127.0.0.1/8"},
			"127.0.0.1",
			true,
			false,
		},
		{
			[]string{"127.0.0.1/8"},
			"127.0.0.2",
			true,
			false,
		},
		{
			[]string{"127.0.0.1/8"},
			"1.1.1.1",
			false,
			false,
		},
		{
			[]string{"127.0.0.1/8", "1.1.1.1/16"},
			"1.1.1.1",
			true,
			false,
		},
		{
			[]string{"127.0.0.1/32", "1.1.1.1/32"},
			"1.1.1.1",
			true,
			false,
		},
	}
	for _, t := range tests {
		ds, err := NewListDataSource(t.data)
		suite.NoError(err)

		ipnet := NewIPNet(ds, "testList")
		err = ipnet.Load()
		suite.NoError(err)

		v, err := ipnet.Check(net.ParseIP(t.check))

		if t.wantErr {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}

		suite.Equal(t.want, v, t)
	}
}

func (suite *ListSuite) Test_CheckIP() {
	tests := []struct {
		data    []string
		check   string
		want    bool
		wantErr bool
	}{
		{
			[]string{"127.0.0.1"},
			"127.0.0.1",
			true,
			false,
		},
		{
			[]string{"127.0.0.1"},
			"127.0.0.2",
			false,
			false,
		},
		{
			[]string{"127.0.0.1"},
			"1.1.1.1",
			false,
			false,
		},
		{
			[]string{"127.0.0.1", "1.1.1.1"},
			"1.1.1.1",
			true,
			false,
		},
	}
	for _, t := range tests {
		ds, err := NewListDataSource(t.data)
		suite.NoError(err)
		f := NewIPNet(ds, "testList")

		err = f.Load()
		suite.NoError(err)

		v, err := f.Check(net.ParseIP(t.check))

		if t.wantErr {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}

		suite.Equal(t.want, v, t)

		// check cache
		v, err = f.Check(net.ParseIP(t.check))

		if t.wantErr {
			suite.Error(err)
		} else {
			suite.NoError(err)
		}

		suite.Equal(t.want, v, t)
	}
}

func TestListSuite(t *testing.T) {
	suite.Run(t, new(ListSuite))
}
