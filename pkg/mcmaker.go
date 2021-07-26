package mcmaker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	ign3types "github.com/coreos/ignition/v2/config/v3_2/types"
	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const roleKey = "machineconfiguration.openshift.io/role"

type McMaker struct {
	name string
	mc   *machineconfigv1.MachineConfig
	i    *ign3types.Config
}

func New(name string) McMaker {
	return McMaker{
		name: name,
		mc: &machineconfigv1.MachineConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: machineconfigv1.GroupVersion.String(),
				Kind:       "MachineConfig",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: make(map[string]string),
			},
			Spec: machineconfigv1.MachineConfigSpec{},
		},
		i: &ign3types.Config{
			Ignition: ign3types.Ignition{
				Version: ign3types.MaxVersion.String(),
			},
		},
	}
}

func (m *McMaker) SetRole(role string) {
	m.mc.ObjectMeta.Name = fmt.Sprintf("%s-%s", m.name, role)
	m.mc.ObjectMeta.Labels[roleKey] = role
}

func normalizeEmpty(src interface{}) interface{} {
	switch t := src.(type) {
	case map[string]interface{}:
		// Recursrvely check the next level down
		t = trimEmptyMap(t)
		// only retain if the trimmed result has content
		if len(t) > 0 {
			return t
		}
		return nil
	case []interface{}:
		// Recursively check all items in the slice
		t = trimEmptySlice(t)
		// only retain if the trimmed result has content
		if len(t) > 0 {
			return t
		}
		return nil
	case string:
		// omit empty strings
		if len(t) == 0 {
			return nil
		}
	case bool:
		// omit false booleans
		if !t {
			return nil
		}
	case int, float64:
		// omit zeroes
		if t == 0 {
			return nil
		}
	case nil:
		// omit nil pointers
		return nil
	default:
		// Report but retain everything else
		fmt.Fprintf(os.Stderr, "Unknown type: %v (%T)\n", src, src)
	}
	return src
}

func trimEmptySlice(src []interface{}) []interface{} {
	var dst []interface{}
	for _, v := range src {
		t := normalizeEmpty(v)
		if t != nil {
			dst = append(dst, t)
		}
	}
	return dst
}

func trimEmptyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		t := normalizeEmpty(v)
		if t != nil {
			dst[k] = t
		}
	}
	return dst
}

func (m *McMaker) AddFile(fname, path string, mode int) error {
	if path == "" {
		return fmt.Errorf("File entries require a path")
	}
	fdata, err := os.Open(fname)
	if err != nil {
		return err
	}
	var encodedBytes bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedBytes)
	_, err = io.Copy(encoder, fdata)
	if err != nil {
		return err
	}
	encoder.Close()
	encodedContent := fmt.Sprintf("data:text/plain;charset=utf-8;base64,%s", encodedBytes.String())

	f := ign3types.File{
		Node: ign3types.Node{
			Path: path,
		},
		FileEmbedded1: ign3types.FileEmbedded1{
			Contents: ign3types.Resource{
				Source: &encodedContent,
			},
			Mode: &mode,
		},
	}
	m.i.Storage.Files = append(m.i.Storage.Files, f)
	return nil
}

func (m *McMaker) AddUnit(source, name string, enable bool) error {
	s, err := os.Open(source)
	if err != nil {
		return err
	}

	var contents bytes.Buffer
	_, err = io.Copy(&contents, s)
	if err != nil {
		return err
	}

	contentString := contents.String()

	if name == "" {
		name = filepath.Base(source)
	}
	u := ign3types.Unit{
		Name:     name,
		Contents: &contentString,
		Enabled:  &enable,
	}
	m.i.Systemd.Units = append(m.i.Systemd.Units, u)
	return nil
}

func (m *McMaker) WriteTo(output io.Writer) (int64, error) {
	//Combine the ingition struct into the mc struct
	rawIgnition, err := json.Marshal(m.i)
	if err != nil {
		return 0, err
	}
	m.mc.Spec.Config = runtime.RawExtension{Raw: rawIgnition}

	// Marshal to json to do 1st-order stripping (omitempty)
	b, err := json.Marshal(m.mc)
	if err != nil {
		return 0, err
	}

	//convert to raw map for 2nd-order stripping
	var c map[string]interface{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return 0, err
	}

	//custom stripping
	d := normalizeEmpty(c)
	if d == nil {
		return 0, fmt.Errorf("empty machineconfig")
	}

	// Finally marshal to yaml and write it out
	yamlBytes, err := yaml.Marshal(d)
	if err != nil {
		return 0, err
	}
	n, err := output.Write(yamlBytes)
	return int64(n), err
}
