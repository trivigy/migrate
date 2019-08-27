package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/trivigy/migrate/v2/types"
)

type DNSSuite struct {
	suite.Suite
}

func (r *DNSSuite) TestDNSCommand() {
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
			DNS{},
			bytes.NewBuffer(nil),
			[]string{"--help"},
			"Controls instance of domain name system resource\n" +
				"\n" +
				"Usage:\n" +
				"  dns [command]\n" +
				"\n" +
				"Available Commands:\n" +
				"  create      Constructs and starts a new instance of this resource.\n" +
				"  destroy     Stops and removes running instance of this resource.\n" +
				"\n" +
				"Flags:\n" +
				"      --help   Show help information.\n" +
				"\n" +
				"Use \"dns [command] --help\" for more information about a command.\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("dns", testCase.buffer, testCase.args)
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

func TestDNSSuite(t *testing.T) {
	suite.Run(t, new(DNSSuite))
}
