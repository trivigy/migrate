package migrations

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/types"
)

func (r *MigrationsSuite) TestUpCommand() {
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
			Up{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"--try"},
			"==> migration \"0.0.1_create-unittest-table\" (up)\n" +
				"CREATE TABLE unittests (value text);\n",
		},
		{
			false, "",
			Up{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "0", "--try"},
			"==> migration \"0.0.1_create-unittest-table\" (up)\n" +
				"CREATE TABLE unittests (value text);\n" +
				"==> migration \"0.0.2_seed-dummy-data\" (up)\n" +
				"INSERT INTO unittests(value) VALUES ('hello'), ('world');\n" +
				"==> migration \"0.0.3_seed-more-dummy-data\" (up)\n" +
				"INSERT INTO unittests(value) VALUES ('here'), ('there');\n",
		},
		{
			false, "",
			Up{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "1"},
			"migration \"0.0.1_create-unittest-table\" successfully applied (up)\n",
		},
		{
			false, "",
			Up{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "0"},
			"migration \"0.0.2_seed-dummy-data\" successfully applied (up)\n" +
				"migration \"0.0.3_seed-more-dummy-data\" successfully applied (up)\n",
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("up", tc.buffer, tc.args)
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
