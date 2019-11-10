package releases

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
			"unittest",
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("describe", tc.buffer, tc.args)
			if err != nil {
				panic(err.Error())
			}

			if tc.output != strings.TrimSpace(tc.buffer.String()) {
				panic(tc.buffer.String())
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}

	uninstall := Uninstall{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver}
	assert.Nil(r.T(), uninstall.Execute("uninstall", bytes.NewBuffer(nil), []string{}))
}
