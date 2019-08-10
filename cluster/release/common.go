package release

import (
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	v1core "k8s.io/api/core/v1"
	v1err "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/trivigy/migrate/v2/config"
)

type common struct{}

func (r common) GetKubeCtl(cfg config.Cluster) (*kubernetes.Clientset, error) {
	kubeConfig, err := cfg.Driver.KubeConfig()
	if err != nil {
		return nil, err
	}

	kubectl, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	_, err = kubectl.CoreV1().
		Namespaces().
		Get(cfg.Namespace, v1meta.GetOptions{})
	if v1err.IsNotFound(err) {
		_, err := kubectl.CoreV1().
			Namespaces().
			Create(&v1core.Namespace{
				TypeMeta: v1meta.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: v1meta.ObjectMeta{
					Name: cfg.Namespace,
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

func (r common) TrimmedYAML(value interface{}) (string, error) {
	rbytesJSON, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	rbytesYAML, err := yaml.JSONToYAML(rbytesJSON)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(rbytesYAML)), nil
}

func (r common) ChunkString(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{s}
	}

	for i := 0; i < len(runes); i += chunkSize {
		nn := i + chunkSize
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}
