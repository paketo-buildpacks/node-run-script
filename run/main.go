package main

import (
	"os"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	environment := noderunscript.Environment{
		NodeRunScripts:  os.Getenv("BP_NODE_RUN_SCRIPTS"),
		NodeProjectPath: os.Getenv("BP_NODE_PROJECT_PATH"),
		LogLevel:        os.Getenv("BP_LOG_LEVEL"),
	}

	packit.Run(
		noderunscript.Detect(environment),
		noderunscript.Build(
			pexec.NewExecutable("npm"),
			pexec.NewExecutable("yarn"),
			chronos.DefaultClock,
			scribe.NewLogger(os.Stdout).WithLevel(environment.LogLevel),
			environment,
		),
	)
}
