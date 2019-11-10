package types

import (
	"encoding/json"
	"fmt"

	"github.com/blang/semver"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// Release defines a collection of kubernetes manifests which can be released
// together as a logical unit.
type Release struct {
	Name      string           `json:"name,omitempty" yaml:"name,omitempty"`
	Version   semver.Version   `json:"version,omitempty" yaml:"version,omitempty"`
	Manifests []runtime.Object `json:"manifests,omitempty" yaml:"manifests,omitempty"`
}

func (r Release) String() string {
	return fmt.Sprintf("%+v", []string{r.Name, r.Version.String()})
}

// UnmarshalJSON defines custom json unmarshalling procedure.
func (r *Release) UnmarshalJSON(b []byte) error {
	obj := map[string]interface{}{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	if name, ok := obj["name"]; ok {
		r.Name = name.(string)
	}

	if version, ok := obj["version"]; ok {
		version, err := semver.Parse(version.(string))
		if err != nil {
			return err
		}
		r.Version = version
	}

	if manifests, ok := obj["manifests"]; ok {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		for _, manifest := range manifests.([]interface{}) {
			rbytes, err := yaml.Marshal(manifest)
			if err != nil {
				return err
			}

			obj, _, err := decode(rbytes, nil, nil)
			if err != nil {
				return err
			}
			r.Manifests = append(r.Manifests, obj)
		}
	}
	return nil
}
