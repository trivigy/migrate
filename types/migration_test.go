package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/blang/semver"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MigrationSuite struct {
	suite.Suite
}

func (r *MigrationSuite) TestMigration_MarshalUnmarshal() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		migration  *Migration
	}{
		{
			false, "",
			&Migration{
				Name: "unittest",
				Tag:  semver.MustParse("0.0.1"),
				Up: []Operation{
					{Query: `unittest`},
				},
				Down: []Operation{
					{Query: `unittest`},
				},
			},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			rbytes, err := json.Marshal(tc.migration)
			if err != nil {
				panic(err.Error())
			}

			actual := &Migration{}
			err = json.Unmarshal(rbytes, actual)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(r.T(), tc.migration, actual)
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}
