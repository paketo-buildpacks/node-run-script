package noderunscript

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

type BuildPlanMetadata struct {
	Build bool `toml:"build"`
}

func Detect(env Environment) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, packageManager, err := ScriptsToRun(filepath.Join(context.WorkingDir, env.NodeProjectPath), env.NodeRunScripts)
		if err != nil {
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
