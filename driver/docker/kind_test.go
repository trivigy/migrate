package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mg8types "github.com/trivigy/migrate/v2/types"
)

type KindSuite struct {
	suite.Suite
	Namespace string `json:"namespace" yaml:"namespace"`
	Driver    interface {
		mg8types.Creator
		mg8types.Destroyer
		mg8types.KubeConfiged
	} `json:"driver" yaml:"driver"`
}

func (r *KindSuite) SetupSuite() {
	docker, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.38"),
	)
	assert.Nil(r.T(), err)
	defer docker.Close()

	ctx := context.Background()
	refStr := "alpine:latest"
	pullOpts := types.ImagePullOptions{}
	reader, err := docker.ImagePull(ctx, refStr, pullOpts)
	assert.Nil(r.T(), err)

	buffer := bytes.NewBuffer(nil)
	_, err = io.Copy(buffer, reader)
	assert.Nil(r.T(), err)

	ctx = context.Background()
	tagged1 := "gcr.io/unittest-12345/unittest:1.0.0"
	err = docker.ImageTag(ctx, refStr, tagged1)
	assert.Nil(r.T(), err)

	ctx = context.Background()
	tagged2 := "gcr.io/testunit-12345/testunit:2.0.0"
	err = docker.ImageTag(ctx, refStr, tagged2)
	assert.Nil(r.T(), err)

	r.Namespace = "unittest"
	r.Driver = KIND{
		Images: []string{tagged1, tagged2},
		Name:   strings.ToLower(randomdata.SillyName()),
	}
}

func (r *KindSuite) TestSetupTearDownCycle() {
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
			[]string{"--help"},
			"",
		},
	}

	for i, testCase := range testCases {
		failMsg := fmt.Sprintf("testCase: %d %v", i, testCase)
		runner := func() {
			err := r.Driver.Create(testCase.buffer)
			if err != nil {
				panic(err)
			}

			err = r.Driver.Destroy(testCase.buffer)
			if err != nil {
				panic(err)
			}
		}

		if testCase.shouldFail {
			assert.PanicsWithValue(r.T(), testCase.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestReleaseSuite(t *testing.T) {
	suite.Run(t, new(KindSuite))
}
