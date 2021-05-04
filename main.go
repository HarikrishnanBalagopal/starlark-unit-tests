package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

func main() {
	if len(os.Args) != 2 {
		log.Errorf("Usage: test_runner path/to/test.star")
		return
	}
	filePath := os.Args[1]
	log.Infof("Running tests for the path %s", filePath)
	starYaml, err := LoadModule()
	must(err)

	// Execute Starlark program in a file.
	thread := &starlark.Thread{Name: "starlark-transform-tests", Load: myload}
	globals, err := starlark.ExecFile(thread, filePath, nil, starYaml)
	must(err)

	// Retrieve a module global.
	runTests, ok := globals["run_tests"]
	if !ok {
		must(fmt.Errorf("run_tests is missing from the test file at path %s . Actual %+v", filePath, globals))
	}

	// Call Starlark function from Go.
	testOutputs, err := starlark.Call(thread, runTests, nil, nil)
	must(err)
	log.Infof("Test outputs: %+v", testOutputs)
	log.Infof("ALL TESTS PASSED")
}
