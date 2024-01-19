package noderunscript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/libnodejs"
)

func ScriptsToRun(workingDir string, nodeRunScripts string) ([]string, string, error) {
	scripts := strings.Split(nodeRunScripts, ",")
	for i := range scripts {
		scripts[i] = strings.TrimSpace(scripts[i])
	}

	packageJSON, err := libnodejs.ParsePackageJSON(workingDir)
	if err != nil {
		return nil, "", err
	}

	var missing []string
	for _, script := range scripts {
		if _, ok := packageJSON.AllScripts[script]; !ok {
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
