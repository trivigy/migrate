package release

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (r *ReleaseSuite) TestListCommand() {
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
			[]string{},
			"",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			command := NewList(r.config)
			err := command.Execute("list", testCase.buffer, testCase.args)
			if err != nil {
				panic(testCase.buffer.String())
			}

			// if testCase.output != testCase.buffer.String() {
			// 	panic(testCase.buffer.String())
			// }
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}
