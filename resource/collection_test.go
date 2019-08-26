package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type CollectionSuite struct {
	suite.Suite
}

func (r *CollectionSuite) TestCollectionCommand() {
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
			Collection{
				"database":   Database{},
				"kubernetes": Kubernetes{},
			},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Usage:\n" +
				"  collection [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  database    SQL database deployment and migrations management tool.\n" +
				"  kubernetes  Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n" +
				"\n" +
				"Use \"collection [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("collection", testCase.buffer, testCase.args)
			if err != nil {
				panic(err.Error())
			}

			if testCase.output != testCase.buffer.String() {
				fmt.Printf("%q\n", testCase.buffer.String())
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

func TestCollectionSuite(t *testing.T) {
	suite.Run(t, new(CollectionSuite))
}
