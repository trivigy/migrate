package database

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/driver/docker"
	"github.com/trivigy/migrate/driver/provider"
)

type DownSuite struct {
	suite.Suite
	name string
}

func (r *DownSuite) SetupTest() {
	r.name = "down"
}

func (r *DownSuite) TestDownCommand() {
	migrations := []Migration{
		{
			Name: "create-unittest-table",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 1},
			Up: []Operation{
				{Query: `CREATE TABLE unittests (value text)`},
			},
			Down: []Operation{
				{Query: `DROP TABLE unittests`},
			},
		},
		{
			Name: "seed-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 2},
			Up: []Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('hello'), ('world')`},
			},
			Down: []Operation{
				{Query: `DELETE FROM unittests WHERE value in ('hello', 'world')`},
			},
		},
		{
			Name: "seed-more-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 3},
			Up: []Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('here'), ('there')`},
			},
			Down: []Operation{
				{Query: `DELETE FROM unittests WHERE value in ('here', 'there')`},
			},
		},
	}

	defaultConfig := map[string]Config{}
	if os.Getenv("CI") != "true" {
		defaultConfig = map[string]Config{
			"default": {
				Migrations: migrations,
				Driver: docker.Postgres{
					RefName: randomdata.SillyName(),
					Version: "9.6",
					DBName:  "unittest",
					User:    "postgres",
				},
			},
		}
	} else {
		defaultConfig = map[string]Config{
			"default": {
				Migrations: migrations,
				Driver: provider.SQLDatabase{
					Dialect:    "postgres",
					DataSource: "host=localhost user=postgres dbname=unittest sslmode=disable",
				},
			},
		}
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
