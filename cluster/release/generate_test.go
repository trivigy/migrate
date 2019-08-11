package release

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/config"
	"github.com/trivigy/migrate/v2/internal/testutils"
)

func (r *ReleaseSuite) TestGenerateCommand() {
	dir, err := ioutil.TempDir(os.TempDir(), "migrate-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		shouldFail bool
		onFail     string
		config     map[string]config.Cluster
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			r.config,
			bytes.NewBuffer(nil),
			[]string{"example", "-d", dir},
		},
		{
			true, "Error: missing \"default\" environment configuration\n",
			map[string]config.Cluster{},
			bytes.NewBuffer(nil),
			[]string{"example"},
		},
		{
			true, "Error: directory \"./not-found\" not found\n",
			r.config,
			bytes.NewBuffer(nil),
			[]string{"example", "-d", "./not-found"},
		},
		{
			true,
			"Error: accepts 1 arg(s), received 0 for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			r.config,
			bytes.NewBuffer(nil),
			[]string{},
		},
		{
			true,
			"Error: invalid argument \"name:wrong\" for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			r.config,
			bytes.NewBuffer(nil),
			[]string{"name:wrong"},
		},
		{
			true,
			"Error: invalid argument \"name:0.0.0-alpha.2+001\" for \"generate\"\n" +
				"\n" +
				"Usage:  generate NAME[:TAG] [flags]\n",
			r.config,
			bytes.NewBuffer(nil),
			[]string{"name:0.0.0-alpha.2+001"},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewGenerate(testCase.config)
			err := command.Execute("generate", testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			empty, err := testutils.IsDirEmpty(dir)
			if err != nil {
				panic(err)
			}

			if empty {
				panic("not empty")
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}
