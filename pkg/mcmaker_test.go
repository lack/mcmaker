package mcmaker

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMcMakerBase(t *testing.T) {
	m := New("test")
	assert.NotNil(t, m)

	expected := `apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  name: test
spec:
  config:
    ignition:
      version: 3.2.0
`

	var result bytes.Buffer
	_, err := m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))
}

func TestAddFile(t *testing.T) {
	m := New("test")
	assert.NotNil(t, m)

	expected := `apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  name: test
spec:
  config:
    ignition:
      version: 3.2.0
    storage:
      files:
      - contents:
          source: data:text/plain;charset=utf-8;base64,W1VuaXRdCkRlc2NyaXB0aW9uPU5vdCBhIHJlYWwgdW5pdAo=
        mode: 420
        path: /some/path
`
	err := m.AddFile("testdata/example.service", "/some/path", 0644)
	assert.NoError(t, err)

	var result bytes.Buffer
	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))
}

func TestAddUnit(t *testing.T) {
	m := New("test")
	assert.NotNil(t, m)

	expected := `apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  name: test
spec:
  config:
    ignition:
      version: 3.2.0
    systemd:
      units:
      - contents: |
          [Unit]
          Description=Not a real unit
        enabled: true
        name: example.service
`

	err := m.AddUnit("testdata/example.service", "", true)
	assert.NoError(t, err)

	var result bytes.Buffer
	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))
}

func TestAddDropin(t *testing.T) {
	m := New("test")
	assert.NotNil(t, m)

	expected := `apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  name: test
spec:
  config:
    ignition:
      version: 3.2.0
    systemd:
      units:
      - dropins:
        - contents: |
            [Unit]
            Description=Not a real drop-in
          name: dropin.conf
        name: test.service
`

	err := m.AddDropin("testdata/dropin.conf", "test.service", "")
	assert.NoError(t, err)

	var result bytes.Buffer
	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))
}
