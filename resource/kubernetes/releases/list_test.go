package releases

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/assert"

	"github.com/trivigy/migrate/v2/types"
)

func (r *ReleasesSuite) TestListCommand() {
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
			List{Namespace: r.Namespace, Releases: r.Releases, Driver: r.Driver},
			bytes.NewBuffer(nil),
			[]string{},
			"+-------------------------+---------+------------+----------+\n" +
				"|          NAME           | VERSION |    KIND    |  STATUS  |\n" +
				"+-------------------------+---------+------------+----------+\n" +
				"| create-unittest-cluster | 0.0.1   | Service    | NotFound |\n" +
				"|                         |         | Deployment | NotFound |\n" +
				"+-------------------------+---------+------------+----------+\n",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := testCase.cmd.Execute("list", testCase.buffer, testCase.args)
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
