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

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/cmd/kind"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha3"
	logutil "sigs.k8s.io/kind/pkg/log"
	"sigs.k8s.io/yaml"

	"github.com/trivigy/migrate/v2/driver"
)

// Kind represents a driver for the Kubernetes IN Docker (sigs.k8s.io/kind)
// project.
type Kind struct {
	Name   string      `json:"name" yaml:"name" validate:"required"`
	Images []string    `json:"images,omitempty" yaml:"images,omitempty"`
	Config interface{} `json:"config,omitempty" yaml:"config,omitempty"`
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

		typeMeta := metav1.TypeMeta{}
		if r.Config != nil {
			cfg, ok := r.Config.(*v1alpha3.Cluster)
			if !ok {
				return fmt.Errorf("unknown config type: %T", r.Config)
			}
			typeMeta = cfg.TypeMeta
		}
		if err := json.Unmarshal(rbytes, &typeMeta); err != nil {
			return err
		}

		// decode specific (apiVersion, kind)
		switch typeMeta.APIVersion {
		case "kind.sigs.k8s.io/v1alpha3":
			if typeMeta.Kind != "Cluster" {
				return fmt.Errorf("unknown kind %s for apiVersion: %s", typeMeta.APIVersion, typeMeta.Kind)
			}
			cfg := &v1alpha3.Cluster{}
			if r.Config != nil {
				if cfg, ok = r.Config.(*v1alpha3.Cluster); !ok {
					return fmt.Errorf("unknown config type: %T", r.Config)
				}
			}
			if err := json.Unmarshal(rbytes, cfg); err != nil {
				return err
			}
			r.Config = cfg
		}
	}

	return nil
}

// Create executes the resource creation process.
func (r Kind) Create(ctx context.Context, out io.Writer) error {
	args := []string{
		"create", "cluster",
		"--name", r.Name,
		"--wait", "5m",
	}
	if r.Config != nil {
		rbytes, err := yaml.Marshal(r.Config)
		if err != nil {
			return err
		}

		tmpfile, err := ioutil.TempFile(os.TempDir(), "kind-*.yaml")
		if err != nil {
			log.Fatal(err)
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
			"load", "docker-image", image,
			"--name", r.Name,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Destroy executes the resource destruction process.
func (r Kind) Destroy(ctx context.Context, out io.Writer) error {
	if err := r.Execute(out, []string{
		"delete", "cluster",
		"--name", r.Name,
	}); err != nil && !strings.HasPrefix(err.Error(), "unknown cluster") {
		return err
	}
	return nil
}

// Execute is a wrapper function to the kind command runner.
func (r Kind) Execute(out io.Writer, args []string) error {
	log.SetOutput(out)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		ForceColors:     logutil.IsTerminal(log.StandardLogger().Out),
	})

	old := os.Stdout
	rd, wr, _ := os.Pipe()
	os.Stdout = wr

	outC := make(chan error)
	go func() {
		if _, err := io.Copy(out, rd); err != nil {
			outC <- err
			return
		}
		if err := rd.Close(); err != nil {
			outC <- err
			return
		}
		outC <- nil
	}()

	defer func() {
		os.Stdout = old
		if err := wr.Close(); err != nil {
			panic(err)
		}

		err := <-outC
		if err != nil {
			panic(err)
		}
		close(outC)
	}()

	cmd := kind.NewCommand()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// Source returns the data source name for the driver.
func (r Kind) Source(ctx context.Context, out io.Writer) error {
	output := bytes.NewBuffer(nil)
	if err := r.Execute(output, []string{
		"get", "kubeconfig-path",
		"--name", r.Name,
	}); err != nil {
		return err
	}

	rbytes, err := ioutil.ReadFile(strings.TrimSpace(output.String()))
	if err != nil {
		return err
	}

	if _, err := out.Write([]byte(string(rbytes) + "\n")); err != nil {
		return err
	}
	return nil
}
