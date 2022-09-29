package noderunscript_test

import (
	"os"
	"path/filepath"
	"testing"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testScripts(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{
			"name": "mypackage",
			"scripts": {
				"build": "mybuildcommand --args",
				"some-script": "somecommand --args",
				"other-script": "othercommand --args"
			}
		}`), 0600)).To(Succeed())
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a list of scripts to run and the package manager used", func() {
		scripts, manager, err := noderunscript.ScriptsToRun(workingDir, "build")
		Expect(err).NotTo(HaveOccurred())
		Expect(scripts).To(Equal([]string{"build"}))
		Expect(manager).To(Equal("npm"))
	})

	context("when a yarn.lock file is present", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "yarn.lock"), nil, 0600)).To(Succeed())
		})

		it("identifies yarn as the package manager", func() {
			scripts, manager, err := noderunscript.ScriptsToRun(workingDir, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(scripts).To(Equal([]string{"build"}))
			Expect(manager).To(Equal("yarn"))
		})
	})

	context("when specific scripts are requested to run", func() {
		it("returns those scripts", func() {
			scripts, manager, err := noderunscript.ScriptsToRun(workingDir, "some-script")
			Expect(err).NotTo(HaveOccurred())
			Expect(scripts).To(Equal([]string{"some-script"}))
			Expect(manager).To(Equal("npm"))
		})
	})

	context("when a list of scripts is requested to run", func() {
		it("returns those scripts", func() {
			scripts, manager, err := noderunscript.ScriptsToRun(workingDir, "some-script, other-script")
			Expect(err).NotTo(HaveOccurred())
			Expect(scripts).To(Equal([]string{"some-script", "other-script"}))
			Expect(manager).To(Equal("npm"))
		})
	})

	context("when a requested script is missing from the package.json", func() {
		it("returns an error", func() {
			_, _, err := noderunscript.ScriptsToRun(workingDir, "missing-script")
			Expect(err).To(MatchError(`could not find script(s) [missing-script] in package.json`))
		})
	})

	context("failure cases", func() {
		context("when the package.json file does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(workingDir, "package.json"))).To(Succeed())
			})

			it("returns an error", func() {
				_, _, err := noderunscript.ScriptsToRun(workingDir, "some-script")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the package.json file is malformed", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "package.json"), []byte("%%%"), 0600)).To(Succeed())
			})

			it("returns an error", func() {
				_, _, err := noderunscript.ScriptsToRun(workingDir, "some-script")
				Expect(err).To(MatchError(ContainSubstring("invalid character '%'")))
			})
		})
	})
}
