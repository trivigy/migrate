package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/resource/database"
	"github.com/trivigy/migrate/v2/resource/database/migrations"
	"github.com/trivigy/migrate/v2/resource/primitive"
	"github.com/trivigy/migrate/v2/types"
)

type DatabaseSuite struct {
	suite.Suite
}

func (r *DatabaseSuite) TestDatabaseCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			true,
			"accepts 1 arg(s), received 0 for \"database\"\n" +
				"\n" +
				"Usage:\n" +
				"  database COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  migrations  Manages the lifecycle of a database migration.\n" +
				"  source      Prints the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
			Database{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"migrations": database.Migrations{
					"generate": migrations.Generate{},
					"up":       migrations.Up{},
					"down":     migrations.Down{},
					"report":   migrations.Report{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Database{
				"create":  primitive.Create{},
				"destroy": primitive.Destroy{},
				"source":  primitive.Source{},
				"migrations": database.Migrations{
					"generate": migrations.Generate{},
					"up":       migrations.Up{},
					"down":     migrations.Down{},
					"report":   migrations.Report{},
				},
			},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"SQL database deployment and migrations management tool\n" +
				"\n" +
				"Usage:\n" +
				"  database COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"  migrations  Manages the lifecycle of a database migration.\n" +
				"  source      Prints the data source name as a connection string.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("database", testCase.buffer, testCase.args)
			if err != nil {
				panic(err.Error())
			}

			if testCase.output != testCase.buffer.String() {
				fmt.Printf("%q\n", testCase.buffer.String())
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
