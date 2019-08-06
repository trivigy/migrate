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

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/driver/docker"
	"github.com/trivigy/migrate/v2/driver/provider"
	"github.com/trivigy/migrate/v2/types"
)

type ReportSuite struct {
	suite.Suite
	name string
}

func (r *ReportSuite) SetupTest() {
	r.name = "report"
}

func (r *ReportSuite) TestReportCommand() {
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

	defaultConfig := map[string]config.Database{}
	if os.Getenv("CI") != "true" {
		defaultConfig = map[string]config.Database{
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
		defaultConfig = map[string]config.Database{
			"default": {
				Migrations: migrations,
				Driver: provider.SQL{
					Dialect:    "postgres",
					DataSource: "host=localhost user=postgres dbname=unittest sslmode=disable",
				},
			},
		}
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
			[]string{},
			"+-------+-----------------------+---------+\n" +
				"|  TAG  |         NAME          | APPLIED |\n" +
				"+-------+-----------------------+---------+\n" +
				"| 0.0.1 | create-unittest-table | pending |\n" +
				"| 0.0.2 | seed-dummy-data       | pending |\n" +
				"| 0.0.3 | seed-more-dummy-data  | pending |\n" +
				"+-------+-----------------------+---------+\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewReport(defaultConfig)
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

func TestReportSuite(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}
