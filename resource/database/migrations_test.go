package database

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type MigrationSuite struct {
	suite.Suite
}

func (r *MigrationSuite) TestDatabaseCommand() {
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
			"accepts 1 arg(s), received 0 for \"migrations\"\n" +
				"\n" +
				"Usage:\n" +
				"  migrations COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  down        Rolls back to the previously applied migrations.\n" +
				"  generate    Adds a new blank migration with increasing version.\n" +
				"  report      Prints which migrations were applied and when.\n" +
				"  up          Executes the next queued migration.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
			Migrations{},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Migrations{},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Manages the lifecycle of a database migration\n" +
				"\n" +
				"Usage:\n" +
				"  migrations COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  down        Rolls back to the previously applied migrations.\n" +
				"  generate    Adds a new blank migration with increasing version.\n" +
				"  report      Prints which migrations were applied and when.\n" +
				"  up          Executes the next queued migration.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("migrations", testCase.buffer, testCase.args)
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

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}
