package migration

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

type MigrationSuite struct {
	suite.Suite
	config map[string]config.Database
}

func (r *MigrationSuite) SetupSuite() {
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

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.Setup(buffer))
}

func (r *MigrationSuite) TearDownSuite() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.TearDown(buffer))
}

func (r *MigrationSuite) TearDownTest() {
	buffer := bytes.NewBuffer(nil)
	down := NewDown(r.config)
	assert.Nil(r.T(), down.Execute("down", buffer, []string{"-l", "0"}))
}

func (r *MigrationSuite) TestDatabaseCommand() {
	defaultConfig := map[string]config.Database{"default": {}}

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
			"Manages the lifecycle of a database migration\n" +
				"\n" +
				"Usage:\n" +
				"  migration [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  down        Rolls back to the previously applied migrations.\n" +
				"  generate    Adds a new blank migration with increasing version.\n" +
				"  report      Prints which migrations were applied and when.\n" +
				"  up          Executes the next queued migration.\n" +
				"\n" +
				"Flags:\n" +
				"  -e, --env ENV   Run with env ENV configurations. (default \"default\")\n" +
				"      --help      Show help information.\n" +
				"\n" +
				"Use \"migration [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewMigration(defaultConfig)
			err := command.Execute("migration", testCase.buffer, testCase.args)
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

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}
