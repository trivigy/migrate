package manifest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ManifestSuite struct {
	suite.Suite
}

func (r *ManifestSuite) TestMustFetch() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		url        string
	}{
		{
			false, "",
			"https://raw.githubusercontent.com/google/metallb/v0.8.1/manifests/metallb.yaml",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			manifests := MustFetch(testCase.url)
			if len(manifests) != 12 {
				panic(len(manifests))
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestManifestSuite(t *testing.T) {
	suite.Run(t, new(ManifestSuite))
}
