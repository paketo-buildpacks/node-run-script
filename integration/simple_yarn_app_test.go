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

func testSimpleYarnApp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a simple yarn app", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds an OCI image for a simple yarn app", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "simple_yarn_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.YarnInstall.Online,
					settings.Buildpacks.NodeRunScript.Online,
				).
				WithEnv(map[string]string{"BP_NODE_RUN_SCRIPTS": "test_script_1,test_script_2"}).
				WithPullPolicy("never").
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Executing build process",
				"    Running 'yarn run test_script_1'",
				MatchRegexp(`      yarn run v\d+\.\d+\.\d+`),
				"      $ echo \"some commands\"",
				"      some commands",
				MatchRegexp(`    Done in \d+\.\d+s\.`),
				"",
				"    Running 'yarn run test_script_2'",
				MatchRegexp(`      yarn run v\d+\.\d+\.\d+`),
				"      $ touch dummyfile.txt",
				MatchRegexp(`    Done in \d+\.\d+s\.`),
				"",
				MatchRegexp(`    Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
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
}
