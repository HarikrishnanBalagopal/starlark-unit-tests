package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/qri-io/starlib/util"
	log "github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"gopkg.in/yaml.v3"
)

type K8sResourceT map[string]interface{}

type entry struct {
	globals starlark.StringDict
	err     error
}

var (
	once       sync.Once
	yamlModule starlark.StringDict
	cache      = map[string]*entry{}
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func myload(_ *starlark.Thread, module string) (starlark.StringDict, error) {
	e, ok := cache[module]
	if e == nil {
		if ok {
			// request for package whose loading is in progress
			return nil, fmt.Errorf("cycle in load graph")
		}

		// Add a placeholder to indicate "load in progress".
		cache[module] = nil

		// Load and initialize the module in a new thread.
		data, err := ioutil.ReadFile(module)
		if err != nil {
			return nil, err
		}
		thread := &starlark.Thread{Name: "exec " + module, Load: myload}
		globals, err := starlark.ExecFile(thread, module, data, nil)
		e = &entry{globals, err}

		// Update the cache.
		cache[module] = e
	}
	return e.globals, e.err
}

// LoadModule loads the base64 module.
// It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		yamlModule = starlark.StringDict{
			"yaml": &starlarkstruct.Module{
				Name: "yaml",
				Members: starlark.StringDict{
					"loads":     starlark.NewBuiltin("loads", Loads),
					"dumps":     starlark.NewBuiltin("dumps", Dumps),
					"load_file": starlark.NewBuiltin("load_file", LoadFile),
				},
			},
		}
	})
	return yamlModule, nil
}

// Loads gets all values from a yaml source
func Loads(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		source starlark.String
		val    interface{}
	)

	err := starlark.UnpackArgs("loads", args, kwargs, "source", &source)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(source.GoString()), &val); err != nil {
		return starlark.None, err
	}

	return util.Marshal(val)
}

// Dumps serializes a starlark object to a yaml string
func Dumps(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		source starlark.Value
	)

	err := starlark.UnpackArgs("dumps", args, kwargs, "source", &source)
	if err != nil {
		return starlark.None, err
	}

	val, err := util.Unmarshal(source)
	if err != nil {
		return starlark.None, err
	}

	data, err := yaml.Marshal(val)
	if err != nil {
		return starlark.None, err
	}

	return starlark.String(string(data)), nil
}

// LoadFile load a yaml file
func LoadFile(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		source starlark.String
		val    interface{}
	)

	err := starlark.UnpackArgs("loads", args, kwargs, "source", &source)
	if err != nil {
		return nil, err
	}

	sourceData, err := ioutil.ReadFile(source.GoString())
	must(err)

	if err := yaml.Unmarshal(sourceData, &val); err != nil {
		return starlark.None, err
	}

	return util.Marshal(val)
}

// GetK8sResourcesFromYaml decodes k8s resources from yaml
func GetK8sResourcesFromYaml(k8sYaml string) ([]K8sResourceT, error) {
	// TODO: split yaml file into multiple resources

	// NOTE: This roundabout method is required to avoid yaml.v3 unmarshalling timestamps into time.Time
	var resourceI interface{}
	if err := yaml.Unmarshal([]byte(k8sYaml), &resourceI); err != nil {
		log.Errorf("Failed to unmarshal k8s yaml. Error: %q", err)
		return nil, err
	}
	resourceJsonBytes, err := json.Marshal(resourceI)
	if err != nil {
		log.Errorf("Failed to marshal the k8s resource into json. K8s resource:\n+%v\nError: %q", resourceI, err)
		return nil, err
	}
	var k8sResource K8sResourceT
	err = json.Unmarshal(resourceJsonBytes, &k8sResource)
	return []K8sResourceT{k8sResource}, err
}
