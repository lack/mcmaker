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

func TestMixedUnits(t *testing.T) {
	var result bytes.Buffer
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
        dropins:
        - contents: |
            [Unit]
            Description=Not a real drop-in
          name: dropin.conf
        enabled: true
        name: example.service
`

	// Test Unit followed by Drop-in
	result.Truncate(0)
	m := New("test")
	assert.NotNil(t, m)
	err := m.AddUnit("testdata/example.service", "", true)
	assert.NoError(t, err)
	err = m.AddDropin("testdata/dropin.conf", "example.service", "")
	assert.NoError(t, err)

	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))

	// Test Drop-in followed by Unit
	result.Truncate(0)
	m = New("test")
	assert.NotNil(t, m)
	err = m.AddDropin("testdata/dropin.conf", "example.service", "")
	assert.NoError(t, err)
	err = m.AddUnit("testdata/example.service", "", true)
	assert.NoError(t, err)

	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))

	// Test 2 drop-ins
	result.Truncate(0)
	m = New("test")
	assert.NotNil(t, m)
	err = m.AddDropin("testdata/dropin.conf", "example.service", "")
	assert.NoError(t, err)
	err = m.AddDropin("testdata/example.service", "example.service", "dropin2.conf")
	assert.NoError(t, err)

	expected = `apiVersion: machineconfiguration.openshift.io/v1
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
        - contents: |
            [Unit]
            Description=Not a real unit
          name: dropin2.conf
        name: example.service
`

	_, err = m.WriteTo(&result)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(result.Bytes()))

	// Test unit contents collision
	m = New("test")
	assert.NotNil(t, m)
	err = m.AddUnit("testdata/example.service", "", true)
	assert.NoError(t, err)
	err = m.AddUnit("testdata/example.service", "", true)
	assert.Error(t, err)

	// Test dropin name collision
	m = New("test")
	assert.NotNil(t, m)
	err = m.AddDropin("testdata/dropin.conf", "example.service", "")
	assert.NoError(t, err)
	err = m.AddDropin("testdata/dropin.conf", "example.service", "")
	assert.Error(t, err)
}
