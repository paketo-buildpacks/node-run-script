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
	npmExec := pexec.NewExecutable("npm")
	yarnExec := pexec.NewExecutable("yarn")
	scriptManager := noderunscript.NewScriptManager()
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		noderunscript.Detect(scriptManager),
		noderunscript.Build(npmExec, yarnExec, scriptManager, chronos.DefaultClock, logger),
	)
}
