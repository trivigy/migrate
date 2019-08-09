package migration

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *MigrationSuite) TestReportCommand() {
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
			command := NewReport(r.config)
			err := command.Execute("report", testCase.buffer, testCase.args)
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
