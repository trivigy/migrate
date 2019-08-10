package migration

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *MigrationSuite) TestDownCommand() {
	up := NewUp(r.config)
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
			command := NewDown(r.config)
			err := command.Execute("down", testCase.buffer, testCase.args)
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