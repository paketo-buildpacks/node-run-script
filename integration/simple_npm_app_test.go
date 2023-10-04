package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSimpleNPMApp(pack occam.Pack, docker occam.Docker) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name   string
			source string

			pullPolicy              = "never"
			extenderBuildStr        = ""
			extenderBuildStrEscaped = ""
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "simple_npm_app"))
			Expect(err).NotTo(HaveOccurred())

			if settings.Extensions.UbiNodejsExtension.Online != "" {
				pullPolicy = "always"
				extenderBuildStr = "[extender (build)] "
				extenderBuildStrEscaped = `\[extender \(build\)\] `
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("when building a simple npm app", func() {
			var (
				image     occam.Image
				container occam.Container
			)

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
				Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
				Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			})

			it("builds an OCI image for a simple npm app", func() {
				var logs fmt.Stringer
				var err error
				image, logs, err = pack.Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.NPMInstall.Online,
						settings.Buildpacks.NodeRunScript.Online,
					).
					WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "test_script_1,test_script_2"}).
					WithPullPolicy(pullPolicy).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStrEscaped, settings.Buildpack.Name)),
					extenderBuildStr+"  Executing build process",
					extenderBuildStr+"    Running 'npm run test_script_1'",
				))
				Expect(logs).To(ContainLines(
					MatchRegexp(extenderBuildStrEscaped+`      > simple_npm_app@\d+\.\d+\.\d+ test_script_1`),
					extenderBuildStr+"      > echo \"some commands\"",
					extenderBuildStr+"      ",
					extenderBuildStr+"      some commands",
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"    Running 'npm run test_script_2'",
					extenderBuildStr+"      ",
					MatchRegexp(extenderBuildStrEscaped+`      > simple_npm_app@\d+\.\d+\.\d+ test_script_2`),
					extenderBuildStr+"      > touch dummyfile.txt",
					extenderBuildStr+"      ",
				))
				Expect(logs).To(ContainLines(
					MatchRegexp(extenderBuildStrEscaped + `    Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				))

				container, err = docker.Container.Run.
					WithCommand("ls -al /workspace/").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring("dummyfile.txt"))
			})
		})

		context("when BP_NODE_RUN_SCRIPTS is explicitly deactivated", func() {
			it("fails detection", func() {
				_, logs, err := pack.WithVerbose().Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.NPMInstall.Online,
						settings.Buildpacks.NodeRunScript.Online,
					).
					WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": ""}).
					WithPullPolicy(pullPolicy).
					Execute(name, source)
				Expect(err).To(HaveOccurred())
				Expect(logs).To(ContainLines(
					`script running has been deactivated: BP_NODE_RUN_SCRIPTS=""`,
				))
			})
		})
	}
}
