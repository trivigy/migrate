package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PsqlDSNSuite struct {
	suite.Suite
}

func (r *PsqlDSNSuite) TestPsqlDSN_Source() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		url        PsqlDSN
		output     string
	}{
		{
			false, "",
			PsqlDSN{
				Host: "localhost",
			},
			"postgres://localhost?sslmode=disable",
		},
		{
			false, "",
			PsqlDSN{
				Host: "localhost",
				Port: "5432",
			},
			"postgres://localhost:5432?sslmode=disable",
		},
		{
			false, "",
			PsqlDSN{
				Host:   "localhost",
				DBName: "mydb",
			},
			"postgres://localhost/mydb?sslmode=disable",
		},
		{
			false, "",
			PsqlDSN{
				Host: "localhost",
				User: "user",
			},
			"postgres://user@localhost?sslmode=disable",
		},
		{
			false, "",
			PsqlDSN{
				Host:     "localhost",
				User:     "user",
				Password: "secret",
			},
			"postgres://user:secret@localhost?sslmode=disable",
		},
		{
			false, "",
			PsqlDSN{
				Host:           "localhost",
				User:           "other",
				DBName:         "otherdb",
				ConnectTimeout: 10,
			},
			"postgres://other@localhost/otherdb?connect_timeout=10&sslmode=disable",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			if testCase.output != testCase.url.Source() {
				panic(testCase.url.Source())
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestPsqlDSNSuite(t *testing.T) {
	suite.Run(t, new(PsqlDSNSuite))
}
