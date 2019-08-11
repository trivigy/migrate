package types

import (
	"fmt"
	"sort"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReleasesSuite struct {
	suite.Suite
}

func (r *ReleasesSuite) TestReleases_Sort() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		releases   *Releases
		order      [][]string
	}{
		{
			false, "",
			&Releases{
				{
					Name:    "b",
					Version: semver.MustParse("2.0.0"),
				},
				{
					Name:    "b",
					Version: semver.MustParse("1.0.0"),
				},
				{
					Name:    "c",
					Version: semver.MustParse("1.0.0"),
				},
				{
					Name:    "a",
					Version: semver.MustParse("1.0.0"),
				},
			},
			[][]string{
				{"a", "1.0.0"},
				{"b", "1.0.0"},
				{"b", "2.0.0"},
				{"c", "1.0.0"},
			},
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			sort.Sort(testCase.releases)
			for i := range testCase.order {
				if testCase.order[i][0] != (*testCase.releases)[i].Name ||
					testCase.order[i][1] != (*testCase.releases)[i].Version.String() {
					panic(*testCase.releases)
				}
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestReleasesSuite(t *testing.T) {
	suite.Run(t, new(ReleasesSuite))
}
