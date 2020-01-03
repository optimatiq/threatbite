package datasource

import (
	"math/rand"
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

	ip, err = ds.Next()
	suite.Error(err)
	suite.Empty(ip)
}

func TestDatasourceSuite(t *testing.T) {
	suite.Run(t, new(DatasourceSuite))
}
