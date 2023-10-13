package noderunscript

import (
	"errors"
	"os"

	"github.com/paketo-buildpacks/libnodejs"
	"github.com/paketo-buildpacks/packit/v2"
)

type BuildPlanMetadata struct {
	Build bool `toml:"build"`
}

func Detect(env Environment) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		if env.NodeRunScripts == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage(`script running has been deactivated: BP_NODE_RUN_SCRIPTS=""`)
		}

		projectDir, err := libnodejs.FindProjectPath(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}
		_, packageManager, err := ScriptsToRun(projectDir, env.NodeRunScripts)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return packit.DetectResult{}, packit.Fail.WithMessage("no package.json file present")
			}

			return packit.DetectResult{}, err
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: BuildPlanMetadata{Build: true},
					},
					{
						Name:     packageManager,
						Metadata: BuildPlanMetadata{Build: true},
					},
					{
						Name:     "node_modules",
						Metadata: BuildPlanMetadata{Build: true},
					},
				},
			},
		}, nil
	}
}
