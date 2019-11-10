package migrations

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/types"
)

func (r *MigrationsSuite) TestDownCommand() {
	up := Up{Migrations: r.Migrations, Driver: r.Driver}
	assert.Nil(r.T(), up.Execute("up", bytes.NewBuffer(nil), []string{"-l", "0"}))

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
			Down{Migrations: r.Migrations, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"--dry-run"},
			"==> migration \"0.0.3_seed-more-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('here', 'there');\n",
		},
		{
			false, "",
			Down{Migrations: r.Migrations, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "0", "--dry-run"},
			"==> migration \"0.0.3_seed-more-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('here', 'there');\n" +
				"==> migration \"0.0.2_seed-dummy-data\" (down)\n" +
				"DELETE FROM unittests WHERE value in ('hello', 'world');\n" +
				"==> migration \"0.0.1_create-unittest-table\" (down)\n" +
				"DROP TABLE unittests;\n",
		},
		{
			false, "",
			Down{Migrations: r.Migrations, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "1"},
			"migration \"0.0.3_seed-more-dummy-data\" successfully removed (down)\n",
		},
		{
			false, "",
			Down{Migrations: r.Migrations, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"-l", "0"},
			"migration \"0.0.2_seed-dummy-data\" successfully removed (down)\n" +
				"migration \"0.0.1_create-unittest-table\" successfully removed (down)\n",
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("down", tc.buffer, tc.args)
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
