package releases

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
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

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := tc.cmd.Execute("list", tc.buffer, tc.args)
			if err != nil {
				panic(err.Error())
			}

			if tc.output != tc.buffer.String() {
				fmt.Printf("%q\n", tc.buffer.String())
				panic(tc.buffer.String())
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}
