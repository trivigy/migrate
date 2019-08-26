package releases

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/types"
)

func (r *ReleasesSuite) TestDescribeCommand() {
	install := Install{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver}
	assert.Nil(r.T(), install.Execute("install", bytes.NewBuffer(nil), []string{}))

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
			Describe{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{"create-unittest-cluster:0.0.1", "Service", "metadata.name"},
			"locker",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("describe", testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			if testCase.output != strings.TrimSpace(testCase.buffer.String()) {
				panic(testCase.buffer.String())
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}

	uninstall := Uninstall{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver}
	assert.Nil(r.T(), uninstall.Execute("uninstall", bytes.NewBuffer(nil), []string{}))
}
