package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/davecgh/go-spew/spew"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha3"

	"github.com/trivigy/migrate/v2/types"
)

type KindSuite struct {
	suite.Suite
	Namespace string
	Driver    interface {
		types.Creator
		types.Destroyer
		types.Sourcer
	}
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
	pullOpts := dtypes.ImagePullOptions{}
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
	r.Driver = Kind{
		Images: []string{tagged1, tagged2},
		Name:   strings.ToLower(randomdata.SillyName()),
	}
}

func (r *KindSuite) TestKind_SetupTearDownCycle() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		buffer     *bytes.Buffer
		output     string
	}{
		{
			false, "",
			bytes.NewBuffer(nil),
			"",
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			err := r.Driver.Create(context.Background(), tc.buffer)
			if err != nil {
				panic(err.Error())
			}

			err = r.Driver.Destroy(context.Background(), tc.buffer)
			if err != nil {
				panic(err.Error())
			}
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func (r *KindSuite) TestKind_MarshalUnmarshal() {
	testCases := []struct {
		shouldFail bool
		onFail     string
		driver     interface {
			types.Creator
			types.Destroyer
			types.Sourcer
		}
	}{
		{
			false, "",
			&Kind{
				Name:   "unittest",
				Images: []string{"unittest1", "unittest2"},
				Config: &v1alpha3.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "kind.sigs.k8s.io/v1alpha3",
					},
					Nodes: []v1alpha3.Node{
						{Role: "control-plane"},
						{Role: "worker"},
						{Role: "worker"},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		failMsg := fmt.Sprintf("test: %d %v", i, spew.Sprint(tc))
		runner := func() {
			rbytes, err := json.Marshal(tc.driver)
			if err != nil {
				panic(err.Error())
			}

			actual := &Kind{}
			err = json.Unmarshal(rbytes, actual)
			if err != nil {
				panic(err.Error())
			}

			assert.EqualValues(r.T(), tc.driver, actual)
		}

		if tc.shouldFail {
			assert.PanicsWithValue(r.T(), tc.onFail, runner, failMsg)
		} else {
			assert.NotPanics(r.T(), runner, failMsg)
		}
	}
}

func TestKindSuite(t *testing.T) {
	suite.Run(t, new(KindSuite))
}
