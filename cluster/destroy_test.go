package cluster

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *ClusterSuite) TestDestroyCommand() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["destroy"].Driver.Setup(buffer))

	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-e", "destroy"},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewDestroy(r.config)
			err := command.Execute("destroy", testCase.buffer, testCase.args)
			if err != nil {
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
