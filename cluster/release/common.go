package release

import (
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
					Labels: map[string]string{
						"name": cfg.Namespace,
					},
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
