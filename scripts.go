package noderunscript

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ScriptsToRun(workingDir string, nodeRunScripts string) ([]string, string, error) {
	scripts := strings.Split(nodeRunScripts, ",")
	for i := range scripts {
		scripts[i] = strings.TrimSpace(scripts[i])
	}

	file, err := os.Open(filepath.Join(workingDir, "package.json"))
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	err = json.NewDecoder(file).Decode(&pkg)
	if err != nil {
		return nil, "", err
	}

	var missing []string
	for _, script := range scripts {
		if _, ok := pkg.Scripts[script]; !ok {
			missing = append(missing, script)
		}
	}
	if len(missing) > 0 {
		return nil, "", fmt.Errorf("could not find script(s) %s in package.json", missing)
	}

	manager := "npm"
	_, err = os.Stat(filepath.Join(workingDir, "yarn.lock"))
	if err == nil {
		manager = "yarn"
	}

	return scripts, manager, nil
}
