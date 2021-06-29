package main

import (
	"github.com/paketo-buildpacks/packit"
	noderunscript "github.com/accrazed/node-run-script"
)

func main() {
	packit.Run(noderunscript.Detect(), noderunscript.Build())
}
