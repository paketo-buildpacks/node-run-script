package noderunscript_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/paketo-buildpacks/node-run-script/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir   string
		workingDir  string
		projectPath string
		cnbDir      string

		build packit.BuildFunc

		timestamp    time.Time
		clock        chronos.Clock
		logger       scribe.Logger
		loggerBuffer *bytes.Buffer
		npmExec      *fakes.Executable
		yarnExec     *fakes.Executable
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{
			"scripts": {
				"build": "echo \"script build running!\"",
				"some-script": "echo \"script some-script running!\""
			}
		}`), 0600)).To(Succeed())

		npmExec = &fakes.Executable{}
		yarnExec = &fakes.Executable{}

		timestamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timestamp
		})

		loggerBuffer = bytes.NewBuffer(nil)
		logger = scribe.NewLogger(loggerBuffer)

		build = noderunscript.Build(npmExec, yarnExec, clock, logger, noderunscript.Environment{
			NodeRunScripts: "build",
		})
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when using npm", func() {
		it("runs npm commands", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(npmExec.ExecuteCall.CallCount).To(Equal(1))
			Expect(npmExec.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "build"}))
			Expect(npmExec.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		})
	})

	context("when using yarn", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0600)).To(Succeed())
		})

		it("runs yarn commands", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(yarnExec.ExecuteCall.CallCount).To(Equal(1))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"run", "build"}))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Dir).To(Equal(workingDir))
		})
	})

	context("when env var $BP_NODE_RUN_SCRIPTS has spaces among commas", func() {
		var executions []pexec.Execution
		it.Before(func() {
			npmExec.ExecuteCall.Stub = func(execution pexec.Execution) error {
				executions = append(executions, execution)
				return nil
			}

			build = noderunscript.Build(npmExec, yarnExec, clock, logger, noderunscript.Environment{
				NodeRunScripts: "build, some-script",
			})
		})

		it("trims the whitespace and successfully detects the scripts", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(executions).To(HaveLen(2))
			Expect(executions[0].Args).To(Equal([]string{"run", "build"}))
			Expect(executions[1].Args).To(Equal([]string{"run", "some-script"}))
		})
	})

	context("when there is a custom project path set", func() {
		it.Before(func() {
			var err error
			projectPath, err = os.MkdirTemp(workingDir, "custom-project-path")
			Expect(err).NotTo(HaveOccurred())
			base := path.Base(projectPath)
			t.Setenv("BP_NODE_PROJECT_PATH", base)

			customPath := filepath.Base(projectPath)

			Expect(os.WriteFile(filepath.Join(workingDir, customPath, "yarn.lock"), nil, 0600)).To(Succeed())
			Expect(fs.Move(filepath.Join(workingDir, "package.json"), filepath.Join(workingDir, customPath, "package.json"))).To(Succeed())

			build = noderunscript.Build(npmExec, yarnExec, clock, logger, noderunscript.Environment{
				NodeRunScripts: "build",
			})
		})

		it("works and runs the correct commands", func() {
			_, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(yarnExec.ExecuteCall.CallCount).To(Equal(1))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Args).To(
				Equal([]string{"run", "build"}))
			Expect(yarnExec.ExecuteCall.Receives.Execution.Dir).To(Equal(projectPath))
		})
	})

	context("failure cases", func() {
		context("when finding the scripts to run fails", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("failed to find scripts to run")))
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the script getting run has an error", func() {
			it.Before(func() {
				npmExec.ExecuteCall.Stub = func(execution pexec.Execution) error {
					_, err := fmt.Fprintln(execution.Stdout, "some stdout output")
					Expect(err).NotTo(HaveOccurred())

					_, err = fmt.Fprintln(execution.Stderr, "some stderr output")
					Expect(err).NotTo(HaveOccurred())

					return fmt.Errorf("some execute error")
				}
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					BuildpackInfo: packit.BuildpackInfo{
						Name:    "Some Buildpack",
						Version: "some-version",
					},
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{},
					},
					Layers: packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError("some execute error"))

				Expect(loggerBuffer.String()).To(ContainSubstring("some stdout output"))
				Expect(loggerBuffer.String()).To(ContainSubstring("some stderr output"))
			})
		})
	})
}
