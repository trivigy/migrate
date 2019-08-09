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

type DatabaseSuite struct {
	suite.Suite
	config map[string]config.Database
}

func (r *DatabaseSuite) SetupSuite() {
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

	r.config = map[string]config.Database{
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
}

func (r *DatabaseSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.TearDown(buffer))
}

func (r *DatabaseSuite) TestDatabaseCommand() {
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
			[]string{"--help"},
			"SQL database deployment and migrations management tool\n" +
				"\n" +
				"Usage:\n" +
				"  database [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Initializes a new instance of a database.\n" +
				"  destroy     Stops a running instance of a database.\n" +
				"  migration   Manages the lifecycle of a database migration.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n" +
				"\n" +
				"Use \"database [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewDatabase(r.config)
			err := command.Execute("database", testCase.buffer, testCase.args)
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
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}
