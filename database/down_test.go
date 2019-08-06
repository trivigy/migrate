package database

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/docker"
	"github.com/trivigy/migrate/v2/types"
)

type DownSuite struct {
	suite.Suite
	name string
}

func (r *DownSuite) SetupTest() {
	r.name = "down"
}

func (r *DownSuite) TestDownCommand() {
	migrations := []types.Migration{
		{
			Name: "create-unittest-table",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 1},
			Up: []types.Operation{
				{Query: `CREATE TABLE unittests (value text)`},
			},
			Down: []types.Operation{
				{Query: `DROP TABLE unittests`},
			},
		},
		{
			Name: "seed-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 2},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('hello'), ('world')`},
			},
			Down: []types.Operation{
				{Query: `DELETE FROM unittests WHERE value in ('hello', 'world')`},
			},
		},
		{
			Name: "seed-more-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 3},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('here'), ('there')`},
			},
			Down: []types.Operation{
				{Query: `DELETE FROM unittests WHERE value in ('here', 'there')`},
			},
		},
	}

	defaultConfig := map[string]config.Database{
		"default": {
			Migrations: migrations,
			Driver: docker.Postgres{
				RefName: strings.ToLower(randomdata.SillyName()),
				Version: "9.6",
				DBName:  "unittest",
				User:    "postgres",
			},
		},
	}

	create := NewCreate(defaultConfig)
	assert.Nil(r.T(), create.Execute("create", bytes.NewBuffer(nil), []string{}))

	up := NewUp(defaultConfig)
	assert.Nil(r.T(), up.Execute("up", bytes.NewBuffer(nil), []string{"-n", "0"}))

	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"--dry-run"},
			"==> migration \"0.0.3_seed-more-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('here', 'there');\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "0", "--dry-run"},
			"==> migration \"0.0.3_seed-more-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('here', 'there');\n" +
				"==> migration \"0.0.2_seed-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('hello', 'world');\n" +
				"==> migration \"0.0.1_create-unittest-table\" (down)\n" +
				"DROP TABLE unittests;\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "1"},
			"migration \"0.0.3_seed-more-dummy-data\" successfully removed (down)\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "0"},
			"migration \"0.0.2_seed-dummy-data\" successfully removed (down)\n" +
				"migration \"0.0.1_create-unittest-table\" successfully removed (down)\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewDown(defaultConfig)
			err := command.Execute(r.name, testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			if testCase.output != testCase.buffer.String() {
				panic(testCase.buffer.String())
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}

	destroy := NewDestroy(defaultConfig)
	assert.Nil(r.T(), destroy.Execute("destroy", bytes.NewBuffer(nil), []string{}))
}

func TestDownSuite(t *testing.T) {
	suite.Run(t, new(DownSuite))
}
