package noderunscript_test

import (
	"testing"

	noderunscript "github.com/paketo-buildpacks/node-run-script"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("returns a parsed environment", func() {
		environment := noderunscript.LoadEnvironment([]string{
			"BP_NODE_RUN_SCRIPTS=some-node-run-scripts-value",
			"LOG_LEVEL=some-log-level-value",
		})

		Expect(environment).To(Equal(noderunscript.Environment{
			LogLevel:       "some-log-level-value",
			NodeRunScripts: "some-node-run-scripts-value",
		}))
	})

	context("when no values are set", func() {
		it("uses the defaults", func() {
			environment := noderunscript.LoadEnvironment([]string{})

			Expect(environment).To(Equal(noderunscript.Environment{
				LogLevel:       "INFO",
				NodeRunScripts: "build",
			}))
		})
	})

	context("when explicit empty values are given", func() {
		it("uses the empty values", func() {
			environment := noderunscript.LoadEnvironment([]string{
				"BP_NODE_RUN_SCRIPTS=",
				"LOG_LEVEL=",
			})

			Expect(environment).To(Equal(noderunscript.Environment{}))
		})
	})
}
