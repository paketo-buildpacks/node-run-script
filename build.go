package noderunscript

import (
	"bytes"
	"os"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) error
}

func Build(npmExec Executable, yarnExec Executable) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger := scribe.NewLogger(os.Stdout)
		buffer := bytes.NewBuffer(nil)
		mainExecutable := npmExec
		execution := pexec.Execution{
			Dir:    context.WorkingDir,
			Args:   []string{"run-script"},
			Stdout: buffer,
			Stderr: buffer,
		}

		packageManager := getPackageManager(context.WorkingDir)

		if packageManager == "yarn" {
			mainExecutable = yarnExec
			execution.Args[0] = "run"
		}

		packageScripts, err := getPackageScripts(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		envRunScripts, exists := os.LookupEnv("BP_NODE_RUN_SCRIPTS")
		if !exists {
			return packit.BuildResult{},
				packit.Fail.WithMessage("environment variable $BP_NODE_RUN_SCRIPTS is not set")
		}

		envScriptNames := strings.Split(envRunScripts, ",")

		var scripts []string
		for _, envScriptName := range envScriptNames {
			if args, exists := packageScripts[envScriptName]; exists {
				scripts = append(scripts, args)
			}
		}

		for _, script := range scripts {

			logger.Subprocess("Running '%s %s %s'", packageManager, execution.Args[0], script)
			args := strings.Split(script, " ")

			execution.Args = execution.Args[:1]
			execution.Args = append(execution.Args, args...)

			err = mainExecutable.Execute(execution)

			if err != nil {
				logger.Subprocess("%s", buffer.String())
				buffer.Reset()
			}
		}

		return packit.BuildResult{}, nil
	}
}
