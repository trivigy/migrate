package provider

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/cmd/kind"
	logutil "sigs.k8s.io/kind/pkg/log"
)

// Kind represents a driver for the Kubernetes IN Docker (sigs.k8s.io/kind)
// project.
type Kind struct {
	Name string
}

// Setup executes the resource creation process.
func (r Kind) Setup(out io.Writer) error {
	if err := r.Execute(out, []string{
		"create", "cluster",
		"--name", r.Name,
		"--wait", "5m",
	}); err != nil {
		return err
	}
	return nil
}

// TearDown executes the resource destruction process.
func (r Kind) TearDown(out io.Writer) error {
	if err := r.Execute(out, []string{
		"delete", "cluster",
		"--name", r.Name,
	}); err != nil &&
		!strings.HasPrefix(err.Error(), "unknown cluster") {
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
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

// KubeConfig returns the content of kubeconfig file.
func (r Kind) KubeConfig() (*rest.Config, error) {
	buffer := bytes.NewBuffer(nil)
	if err := r.Execute(buffer, []string{
		"get", "kubeconfig-path",
		"--name", r.Name,
	}); err != nil {
		return nil, err
	}

	kubeConfigBytes, err := ioutil.ReadFile(strings.TrimSpace(buffer.String()))
	if err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
	if err != nil {
		return nil, err
	}

	return kubeConfig, nil
}
