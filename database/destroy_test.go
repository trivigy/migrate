package database

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *DatabaseSuite) TestDestroyCommand() {
	buffer := bytes.NewBuffer(nil)
	assert.Nil(r.T(), r.config["default"].Driver.Setup(buffer))

	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			[]string{},
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
