// Package releases implements the releases subcommand structure.
package releases

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	v1core "k8s.io/api/core/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/trivigy/migrate/v2/types"
)

// GetK8sClientset defines a function which generates a new client connection
// to kubernetes.
func GetK8sClientset(ctx context.Context, driver types.Sourcer, namespace string) (*kubernetes.Clientset, error) {
	output := bytes.NewBuffer(nil)
	if err := driver.Source(ctx, output); err != nil {
		return nil, err
	}

	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig(output.Bytes())
	if err != nil {
		return nil, err
	}

	kubectl, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	if namespace == "" {
		namespace = "default"
	}

	_, err = kubectl.CoreV1().
		Namespaces().
		Get(namespace, v1meta.GetOptions{})
	if v1err.IsNotFound(err) {
		_, err := kubectl.CoreV1().
			Namespaces().
			Create(&v1core.Namespace{
				TypeMeta: v1meta.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: v1meta.ObjectMeta{
					Name: namespace,
				},
			})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return kubectl, nil
}

// FallBackNS defines a helper function for ensuring namespace fallback.
func FallBackNS(namespace, fallback string) string {
	if namespace != "" {
		return namespace
	}

	if fallback != "" {
		return fallback
	}

	return "default"
}

// NewEmbeddedTable defines helper function for generating tabled data.
func NewEmbeddedTable() (*tablewriter.Table, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	tbl := tablewriter.NewWriter(buf)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetBorder(false)
	tbl.SetAutoWrapText(false)
	return tbl, buf
}

// FilterResult helper function allows jq like filtering through a json.
func FilterResult(value interface{}, filter string) (string, error) {
	rbytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	found := string(rbytes)
	if filter != "" && filter != "." {
		found = gjson.Get(string(rbytes), filter).Raw
	}

	result, err := yaml.JSONToYAML([]byte(found))
	if err != nil {
		return "", err
	}

	return string(result), nil
}
