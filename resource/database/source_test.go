package database

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/internal/testutils"
	"github.com/trivigy/migrate/v2/types"
)

type SourceSuite struct {
	suite.Suite
}

func (r *SourceSuite) TestSourceCommand() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		cmd        types.Command
		buffer     *bytes.Buffer
		args       []string
	}{
		{
			true, "implement me",
			Source{Driver: testutils.Driver{}},
			bytes.NewBuffer(nil),
			[]string{},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("create", testCase.buffer, testCase.args)
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

func TestSourceSuite(t *testing.T) {
	suite.Run(t, new(SourceSuite))
}
