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

func testVueYarnApp(pack occam.Pack, docker occam.Docker) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually
		)

		context("when building a Vue yarn app", func() {
			var (
				image     occam.Image
				container occam.Container

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

				if settings.Extensions.UbiNodejsExtension.Online != "" {
					pullPolicy = "always"
					extenderBuildStr = "[extender (build)] "
					extenderBuildStrEscaped = `\[extender \(build\)\] `
				}
			})

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
				Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
				Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
				Expect(os.RemoveAll(source)).To(Succeed())
			})

			it("builds an OCI image for a Vue yarn app", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "vue_yarn_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.Yarn.Online,
						settings.Buildpacks.YarnInstall.Online,
						settings.Buildpacks.NodeRunScript.Online,
					).
					WithPullPolicy(pullPolicy).
					Execute(name, source)

				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStrEscaped, settings.Buildpack.Name)),
					extenderBuildStr+"  Executing build process",
					extenderBuildStr+"    Running 'yarn run build'",
					MatchRegexp(extenderBuildStrEscaped+`      yarn run v\d+\.\d+\.\d+`),
					extenderBuildStr+"      $ vue-cli-service build",
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr + "       DONE  Build complete. The dist directory is ready to be deployed.",
				))

				Expect(logs).To(ContainLines(MatchRegexp(extenderBuildStrEscaped + `    Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`)))

				container, err = docker.Container.Run.
					WithCommand("ls -al /workspace/dist/").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring("index.html"))
			})
		})
	}
}
