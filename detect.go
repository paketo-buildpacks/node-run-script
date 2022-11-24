package noderunscript

import (
	"errors"
	"os"
	"path/filepath"

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

		_, packageManager, err := ScriptsToRun(filepath.Join(context.WorkingDir, env.NodeProjectPath), env.NodeRunScripts)
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
