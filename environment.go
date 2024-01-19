package noderunscript

import (
	"strings"
)

type Environment struct {
	LogLevel       string
	NodeRunScripts string
}

func LoadEnvironment(variables []string) Environment {
	environment := Environment{
		LogLevel:       "INFO",
		NodeRunScripts: "build",
	}

	for _, variable := range variables {
		key, value, found := strings.Cut(variable, "=")
		if found {
			switch key {
			case "LOG_LEVEL":
				environment.LogLevel = value
			case "BP_NODE_RUN_SCRIPTS":
				environment.NodeRunScripts = value
			}
		}
	}

	return environment
}
