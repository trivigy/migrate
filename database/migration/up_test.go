package migration

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *MigrationSuite) TestUpCommand() {
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
			command := NewUp(r.config)
			err := command.Execute("up", testCase.buffer, testCase.args)
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
