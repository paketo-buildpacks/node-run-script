package noderunscript_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNodeRunScript(t *testing.T) {
	suite := spec.New("node-run-script", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("Environment", testEnvironment)
	suite("Scripts", testScripts)
	suite.Run(t)
}
