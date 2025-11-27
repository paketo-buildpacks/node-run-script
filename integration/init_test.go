package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/format"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpacks struct {
		NodeEngine struct {
			Online string
		}

		NPMInstall struct {
			Online string
		}

		Yarn struct {
			Online string
		}

		YarnInstall struct {
			Online string
		}

		NodeRunScript struct {
			Online string
		}
	}
	Extensions struct {
		UbiNodejsExtension struct {
			Online string
		}
	}
	Buildpack struct {
		ID   string
		Name string
	}
	Config struct {
		NodeEngine         string `json:"node-engine"`
		NPMInstall         string `json:"npm-install"`
		Yarn               string `json:"yarn"`
		YarnInstall        string `json:"yarn-install"`
		UbiNodejsExtension string `json:"ubi-nodejs-extension"`
	}
}

func TestIntegration(t *testing.T) {
	format.MaxLength = 0
	Expect := NewWithT(t).Expect
	SetDefaultEventuallyTimeout(10 * time.Second)

	integrationFile, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())
	Expect(json.NewDecoder(integrationFile).Decode(&settings.Config)).To(Succeed())
	Expect(integrationFile.Close()).To(Succeed())

	buildpackFile, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	_, err = toml.NewDecoder(buildpackFile).Decode(&settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpackFile.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	pack := occam.NewPack()

	builder, err := pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	if builder.BuilderName == "paketobuildpacks/builder-ubi8-buildpackless-base" {
		settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
			Execute(settings.Config.UbiNodejsExtension)
		Expect(err).ToNot(HaveOccurred())
	}

	settings.Buildpacks.NodeRunScript.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NPMInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.NPMInstall)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Yarn.Online, err = buildpackStore.Get.
		Execute(settings.Config.Yarn)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.YarnInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.YarnInstall)
	Expect(err).NotTo(HaveOccurred())

	docker := occam.NewDocker()

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("ProjectPathApp", testProjectPathApp(pack, docker))
	suite("SimpleNPMApp", testSimpleNPMApp(pack, docker))
	suite("SimpleYarnApp", testSimpleYarnApp(pack, docker))
	suite("VueNPMApp", testVueNPMApp(pack, docker))
	suite("VueYarnApp", testVueYarnApp(pack, docker))
	suite.Run(t)
}
