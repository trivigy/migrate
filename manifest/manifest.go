package manifest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	v1beta1policy "k8s.io/api/policy/v1beta1"
	v1rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// MustFetch fetches kubernetes manifests online or panics upon error.
func MustFetch(urls ...string) []interface{} {
	manifests := make([]interface{}, 0)
	for _, url := range urls {
		rbytes, err := FetchYAML(url)
		if err != nil {
			panic(err)
		}

		results, err := ParseYAML(rbytes)
		if err != nil {
			panic(err)
		}
		manifests = append(manifests, results...)
	}
	return manifests
}

// FetchYAML fetches an individual kubernetes manifest.
func FetchYAML(url string) ([]byte, error) {
	out := bytes.NewBuffer(nil)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf(resp.Status))
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), err
}

// ParseYAML converts a yaml manifets file into runtime manifest objects.
func ParseYAML(rbytes []byte) ([]interface{}, error) {
	fileAsString := string(rbytes[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	retVal := make([]interface{}, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)
		if err != nil {
			return nil, err
		}

		switch obj := obj.(type) {
		case *v1core.Namespace,
			*v1core.Pod,
			*v1core.ServiceAccount,
			*v1core.ConfigMap,
			*v1rbac.Role,
			*v1rbac.RoleBinding,
			*v1rbac.ClusterRole,
			*v1rbac.ClusterRoleBinding,
			*v1beta1policy.PodSecurityPolicy,
			*v1apps.DaemonSet,
			*v1apps.Deployment:
			retVal = append(retVal, obj)
		default:
			panic(fmt.Errorf("element type %T not supported", obj))
		}
	}
	return retVal, nil
}
