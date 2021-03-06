package releases

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/testutils"
	"github.com/trivigy/migrate/v2/types"
)

func (r *ReleasesSuite) TestGenerateCommand() {
	dir, err := ioutil.TempDir(os.TempDir(), "migrate-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			Generate{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"example", "-d", dir},
		},
		{
			true, "directory \"./not-found\" not found",
			Generate{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"example", "-d", "./not-found"},
		},
		{
			true,
			"accepts 1 arg(s), received 0 for \"generate\"\n" +
				"\n" +
				"Usage:\n" +
				"  generate NAME[:TAG] [flags]\n" +
				"\n" +
				"Flags:\n" +
				"  -d, --dir PATH   Specify directory PATH where to generate miration file. (default \".\")\n" +
				"      --help       Show help information.\n",
			Generate{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{},
		},
		{
			true,
			"invalid argument \"name:wrong\" for \"generate\"\n" +
				"\n" +
				"Usage:\n" +
				"  generate NAME[:TAG] [flags]\n" +
				"\n" +
				"Flags:\n" +
				"  -d, --dir PATH   Specify directory PATH where to generate miration file. (default \".\")\n" +
				"      --help       Show help information.\n",
			Generate{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"name:wrong"},
		},
		{
			true,
			"invalid argument \"name:0.0.0-alpha.2+001\" for \"generate\"\n" +
				"\n" +
				"Usage:\n" +
				"  generate NAME[:TAG] [flags]\n" +
				"\n" +
				"Flags:\n" +
				"  -d, --dir PATH   Specify directory PATH where to generate miration file. (default \".\")\n" +
				"      --help       Show help information.\n",
			Generate{Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"name:0.0.0-alpha.2+001"},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("generate", tc.buffer, tc.args)
			if err != nil {
				panic(err.Error())
			}

			empty, err := testutils.IsDirEmpty(dir)
			if err != nil {
				panic(err)
			}

			if empty {
				panic("not empty")
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}
