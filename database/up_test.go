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

type UpSuite struct {
	suite.Suite
	name string
}

func (r *UpSuite) SetupTest() {
	r.name = "up"
}

func (r *UpSuite) TestUpCommand() {
	migrations := []types.Migration{
		{
			Name: "create-unittest-table",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 1},
			Up: []types.Operation{
				{Query: `CREATE TABLE unittests (value text)`},
			},
		},
		{
			Name: "seed-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 2},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('hello'), ('world')`},
			},
		},
		{
			Name: "seed-more-dummy-data",
			Tag:  semver.Version{Major: 0, Minor: 0, Patch: 3},
			Up: []types.Operation{
				{Query: `INSERT INTO unittests(value) VALUES ('here'), ('there')`},
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
			"==> migration \"0.0.1_create-unittest-table\" (up)\n" +
				"CREATE TABLE unittests (value text);\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "0", "--dry-run"},
			"==> migration \"0.0.1_create-unittest-table\" (up)\n" +
				"CREATE TABLE unittests (value text);\n" +
				"==> migration \"0.0.2_seed-dummy-data\" (up)\n" +
				"INSERT INTO unittests(value) VALUES ('hello'), ('world');\n" +
				"==> migration \"0.0.3_seed-more-dummy-data\" (up)\n" +
				"INSERT INTO unittests(value) VALUES ('here'), ('there');\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "1"},
			"migration \"0.0.1_create-unittest-table\" successfully applied (up)\n",
		},
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-n", "0"},
			"migration \"0.0.2_seed-dummy-data\" successfully applied (up)\n" +
				"migration \"0.0.3_seed-more-dummy-data\" successfully applied (up)\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewUp(defaultConfig)
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

func TestUpSuite(t *testing.T) {
	suite.Run(t, new(UpSuite))
}
