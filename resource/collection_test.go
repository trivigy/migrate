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
			true,
			"accepts 1 arg(s), received 0 for \"collection\"\n" +
				"\n" +
				"Usage:\n" +
				"  collection COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  database    SQL database deployment and migrations management tool.\n" +
				"  domainName  Controls instance of domain name service resource.\n" +
				"  kubernetes  Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
			Collection{
				"database":   Database{},
				"kubernetes": Kubernetes{},
				"domainName": DomainName{},
			},
			bytes.NewBuffer(nil),
			[]string{},
			"",
		},
		{
			false, "",
			Collection{
				"database":   Database{},
				"kubernetes": Kubernetes{},
				"domainName": DomainName{},
			},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Usage:\n" +
				"  collection COMMAND [flags]\n" +
				"\n" +
				"Available Commands:\n" +
				"  database    SQL database deployment and migrations management tool.\n" +
				"  domainName  Controls instance of domain name service resource.\n" +
				"  kubernetes  Kubernetes cluster release and deployment controller.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n",
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
