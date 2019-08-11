package cluster

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *ClusterSuite) TestCreateCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{"-e", "create"},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewCreate(r.config)
			err := command.Execute("create", testCase.buffer, testCase.args)
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

	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["create"].Driver.TearDown(buffer))
}
