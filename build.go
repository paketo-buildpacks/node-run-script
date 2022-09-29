package noderunscript

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) error
}

func Build(npm Executable, yarn Executable, clock chronos.Clock, logger scribe.Logger, env Environment) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		projectDir := filepath.Join(context.WorkingDir, env.NodeProjectPath)
		scripts, packageManager, err := ScriptsToRun(projectDir, env.NodeRunScripts)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to find scripts to run: %w", err)
		}

		exec := npm
		if packageManager == "yarn" {
			exec = yarn
		}

		logger.Process("Executing build process")
		duration, err := clock.Measure(func() error {
			for _, script := range scripts {
				logger.Subprocess("Running '%s %s %s'", packageManager, "run", script)

				err := exec.Execute(pexec.Execution{
					Dir:    projectDir,
					Args:   []string{"run", script},
					Stdout: logger.ActionWriter,
					Stderr: logger.ActionWriter,
				})
				if err != nil {
					return err
				}

				logger.Break()
			}

			return nil
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		return packit.BuildResult{}, nil
	}
}
