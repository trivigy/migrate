package docker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type PostgresSuite struct {
	suite.Suite
}

func (r *PostgresSuite) TestPostgres_MarshalUnmarshal() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		driver     interface {
			types.Creator
			types.Destroyer
			types.Sourcer
		}
	}{
		{
			false, "",
			&Postgres{
				Name:         "unittest",
				Version:      "unittest",
				User:         "unittest",
				Password:     "unittest",
				DBName:       "unittest",
				InitDBArgs:   "unittest",
				InitDBWalDir: "unittest",
				PGData:       "unittest",
			},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			rbytes, err := json.Marshal(tc.driver)
			if err != nil {
				panic(err.Error())
			}

			actual := &Postgres{}
			err = json.Unmarshal(rbytes, actual)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(r.T(), tc.driver, actual)
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
