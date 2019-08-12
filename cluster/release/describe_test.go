package release

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert"
)

func (r *ReleaseSuite) TestDescribeCommand() {
	install := NewInstall(r.config)
	assert.Nil(r.T(), install.Execute("install", bytes.NewBuffer(nil), []string{}))

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
			[]string{"create-unittest-cluster:0.0.1", "Service", "metadata.name"},
			"locker",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewDescribe(r.config)
			err := command.Execute("describe", testCase.buffer, testCase.args)
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

	uninstall := NewUninstall(r.config)
	assert.Nil(r.T(), uninstall.Execute("uninstall", bytes.NewBuffer(nil), []string{}))
}
