package noderunscript

import "strings"

type Environment struct {
	LogLevel        string
	NodeProjectPath string
	NodeRunScripts  string
}

func LoadEnvironment(variables []string) Environment {
	environment := Environment{
		LogLevel:        "INFO",
		NodeProjectPath: ".",
		NodeRunScripts:  "build",
	}

	for _, variable := range variables {
		key, value, found := strings.Cut(variable, "=")
		if found {
			switch key {
			case "LOG_LEVEL":
				environment.LogLevel = value
			case "BP_NODE_PROJECT_PATH":
				environment.NodeProjectPath = value
			case "BP_NODE_RUN_SCRIPTS":
				environment.NodeRunScripts = value
			}
		}
	}

	return environment
}
