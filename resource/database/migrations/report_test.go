package migrations

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/types"
)

func (r *MigrationsSuite) TestReportCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
		output     string
	}{
		{
			false, "",
			Report{Driver: r.Driver},
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

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("report", tc.buffer, tc.args)
			if err != nil {
				panic(err.Error())
			}

			if tc.output != tc.buffer.String() {
				panic(tc.buffer.String())
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}
