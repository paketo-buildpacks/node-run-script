package noderunscript_test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{
			"scripts": {
				"build": "mybuildcommand --args",
				"some-script": "somecommand --args"
			}
		}`), 0600)).To(Succeed())

		detect = noderunscript.Detect(noderunscript.Environment{
			NodeRunScripts: "build",
		})
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when using npm", func() {
		it("returns a plan that requires node and npm", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "npm",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "node_modules",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
				},
			}))
		})
	})

	context("when using yarn", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0600)).To(Succeed())
		})

		it("returns a plan that requires node and yarn", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name:     "node",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "yarn",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
					{
						Name:     "node_modules",
						Metadata: noderunscript.BuildPlanMetadata{Build: true},
					},
				},
			}))
		})
	})

	context("when env var $BP_NODE_RUN_SCRIPTS has spaces among commas", func() {
		it.Before(func() {
			detect = noderunscript.Detect(noderunscript.Environment{NodeRunScripts: "build, some-script"})
		})

		it("trims the whitespace and successfully detects the scripts", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("when env var $BP_NODE_RUN_SCRIPTS is empty", func() {
		it.Before(func() {
			detect = noderunscript.Detect(noderunscript.Environment{NodeRunScripts: ""})
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage(`script running has been deactivated: BP_NODE_RUN_SCRIPTS=""`)))
		})
	})

	context("when env var $BP_NODE_PROJECT_PATH is set", func() {
		it.Before(func() {
			customPath, err := os.MkdirTemp(workingDir, "custom-project-path")
			base := path.Base(customPath)
			t.Setenv("BP_NODE_PROJECT_PATH", base)
			Expect(err).NotTo(HaveOccurred())
			customPath = filepath.Base(customPath)

			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "yarn.lock"), nil, 0600)).To(Succeed())
			Expect(fs.Move(filepath.Join(workingDir, "package.json"), filepath.Join(workingDir, customPath, "package.json"))).To(Succeed())

			detect = noderunscript.Detect(noderunscript.Environment{
				NodeRunScripts: "build",
			})
		})

		context("when the custom project path contains yarn.lock", func() {
			it("returns a plan that requires node and yarn", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Requires: []packit.BuildPlanRequirement{
						{
							Name:     "node",
							Metadata: noderunscript.BuildPlanMetadata{Build: true},
						},
						{
							Name:     "yarn",
							Metadata: noderunscript.BuildPlanMetadata{Build: true},
						},
						{
							Name:     "node_modules",
							Metadata: noderunscript.BuildPlanMetadata{Build: true},
						},
					},
				}))
			})
		})
	})

	context("failure cases", func() {
		context("if package.json is absent", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError(packit.Fail.WithMessage("no package.json file present")))
			})
		})

		context("if any of the scripts in \"$BP_NODE_RUN_SCRIPTS\" does not exist in package.json", func() {
			it.Before(func() {
				detect = noderunscript.Detect(noderunscript.Environment{
					NodeRunScripts: "build,script1,some-script,script2,script3",
				})
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError("could not find script(s) [script1 script2 script3] in package.json"))
			})
		})

		context("if $BP_NODE_PROJECT_PATH leads to a directory that doesn't exist", func() {
			it.Before(func() {
				detect = noderunscript.Detect(noderunscript.Environment{
					NodeRunScripts: "build",
				})
				t.Setenv("BP_NODE_PROJECT_PATH", "not_a_real_directory")
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(HaveOccurred())
			})
		})
	})
}
