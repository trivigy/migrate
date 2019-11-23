package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kind/cmd/kind/app"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"

	"github.com/trivigy/migrate/v2/driver"
)

// Kind represents a driver for the Kubernetes IN Docker (sigs.k8s.io/kind)
// project.
type Kind struct {
	Name   string            `json:"name" yaml:"name"`
	Images []string          `json:"images,omitempty" yaml:"images,omitempty"`
	Config *v1alpha4.Cluster `json:"config,omitempty" yaml:"config,omitempty"`
}

var _ interface {
	driver.WithCreate
	driver.WithDestroy
	driver.WithSource
} = new(Kind)

// UnmarshalJSON defines custom json unmarshalling procedure.
func (r *Kind) UnmarshalJSON(b []byte) error {
	obj := map[string]interface{}{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	if name, ok := obj["name"]; ok {
		r.Name = name.(string)
	}

	if images, ok := obj["images"]; ok {
		convert := make([]string, len(images.([]interface{})))
		for i, image := range images.([]interface{}) {
			convert[i] = image.(string)
		}
		r.Images = convert
	}

	if config, ok := obj["config"]; ok {
		rbytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		typeMeta := v1alpha4.TypeMeta{}
		if r.Config != nil {
			typeMeta = r.Config.TypeMeta
		}
		if err := json.Unmarshal(rbytes, &typeMeta); err != nil {
			return err
		}

		// decode specific (apiVersion, kind)
		switch typeMeta.APIVersion {
		case "kind.x-k8s.io/v1alpha4":
			if typeMeta.Kind != "Cluster" {
				return fmt.Errorf("unknown kind %s for apiVersion: %s", typeMeta.APIVersion, typeMeta.Kind)
			}
			cfg := &v1alpha4.Cluster{}
			if r.Config != nil {
				cfg = r.Config
			}
			if err := json.Unmarshal(rbytes, cfg); err != nil {
				return err
			}
			r.Config = cfg
		default:
			return fmt.Errorf("version not supported %q", typeMeta.APIVersion)
		}
	}

	return nil
}

// Create executes the resource creation process.
func (r Kind) Create(ctx context.Context, out io.Writer) error {
	args := []string{"create", "cluster", "--name", r.Name, "--wait", "5m"}
	if r.Config != nil {
		rbytes, err := yaml.Marshal(r.Config)
		if err != nil {
			return err
		}

		tmpfile, err := ioutil.TempFile(os.TempDir(), "kind-*.yaml")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write(rbytes); err != nil {
			return err
		}
		if err := tmpfile.Close(); err != nil {
			return err
		}

		args = append(args, "--config", tmpfile.Name())
	}

	if err := r.Execute(out, args); err != nil {
		return err
	}

	for _, image := range r.Images {
		if err := r.Execute(out, []string{
			"load", "docker-image", image, "--name", r.Name,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Destroy executes the resource destruction process.
func (r Kind) Destroy(ctx context.Context, out io.Writer) error {
	if err := r.Execute(out, []string{
		"delete", "cluster", "--name", r.Name,
	}); err != nil && !strings.HasPrefix(err.Error(), "unknown cluster") {
		return err
	}
	return nil
}

// Execute is a wrapper function to the kind command runner.
func (r Kind) Execute(out io.Writer, args []string) error {
	logger := cmd.NewLogger()
	setWriter(logger, out)
	streams := cmd.IOStreams{In: os.Stdin, Out: out, ErrOut: out}
	if err := app.Run(logger, streams, args); err != nil {
		return err
	}
	return nil
}

// setWriter will call logger.SetWriter(w) if logger has a SetWriter method
func setWriter(logger log.Logger, w io.Writer) {
	type writerSetter interface {
		SetWriter(io.Writer)
	}
	v, ok := logger.(writerSetter)
	if ok {
		v.SetWriter(w)
	}
}

// Source returns the data source name for the driver.
func (r Kind) Source(ctx context.Context, out io.Writer) error {
	output := bytes.NewBuffer(nil)
	if err := r.Execute(output, []string{
		"get", "kubeconfig", "--name", r.Name,
	}); err != nil {
		return err
	}

	if _, err := out.Write(output.Bytes()); err != nil {
		return err
	}
	return nil
}
