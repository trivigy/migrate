// Package releases implements the releases subcommand structure.
package releases

import (
	"bytes"
	"context"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	v1core "k8s.io/api/core/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"

	"github.com/trivigy/migrate/v2/driver"
)

// GetK8sClientset defines a function which generates a new client connection
// to kubernetes.
func GetK8sClientset(ctx context.Context, driver driver.WithSource, namespace string) (*kubernetes.Clientset, error) {
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
		Get(ctx, namespace, v1meta.GetOptions{})
	if v1err.IsNotFound(err) {
		_, err := kubectl.CoreV1().
			Namespaces().
			Create(ctx, &v1core.Namespace{
				TypeMeta: v1meta.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: v1meta.ObjectMeta{
					Name: namespace,
				},
			}, v1meta.CreateOptions{})
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

// EmbeddedTable represents an in memory data aggregator for a single kube
// manifest kind.
type EmbeddedTable struct {
	Results []runtime.Object
	Table   *tablewriter.Table
	Buffer  *bytes.Buffer
}

// NewEmbeddedTable defines helper function for generating tabled data.
func NewEmbeddedTable() *EmbeddedTable {
	tbl := &EmbeddedTable{Buffer: bytes.NewBuffer(nil)}
	tbl.Table = tablewriter.NewWriter(tbl.Buffer)
	tbl.Table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.Table.SetBorder(false)
	tbl.Table.SetAutoWrapText(false)
	return tbl
}

// FilterResult helper function allows jq like filtering through a json.
func FilterResult(value interface{}, filter string) (string, error) {
	rbytes, err := yaml.Marshal(value)
	if err != nil {
		return "", err
	}

	found, err := yaml.YAMLToJSON(rbytes)
	if err != nil {
		return "", err
	}

	if filter != "" && filter != "." {
		found = []byte(gjson.Get(string(found), filter).Raw)
	}

	result, err := yaml.JSONToYAML(found)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
