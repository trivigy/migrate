package release

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	v1core "k8s.io/api/core/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/trivigy/migrate/v2/config"
)

type common struct{}

func (r common) GetKubeCtl(cfg *config.Cluster) (*kubernetes.Clientset, error) {
	kubeConfig, err := cfg.Driver.KubeConfig()
	if err != nil {
		return nil, err
	}

	kubectl, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	namespace := "default"
	if cfg.Namespace != "" {
		namespace = cfg.Namespace
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

func (r common) Namespace(cfg config.Cluster, namespace string) string {
	if namespace != "" {
		return namespace
	}

	if cfg.Namespace != "" {
		return cfg.Namespace
	}

	return "default"
}

func (r common) NewEmbeddedTable() (*tablewriter.Table, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	tbl := tablewriter.NewWriter(buf)
	tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tbl.SetBorder(false)
	tbl.SetAutoWrapText(false)
	return tbl, buf
}

func (r common) FilterResult(value interface{}, filter string) (string, error) {
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
