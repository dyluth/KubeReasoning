package kubeloader

import (
	"fmt"
	"strings"
)

func GetClusters(filter string) ([]string, error) {

	out, err := runGetStdout("kubectl", "config", "get-contexts", "-o=name")
	if err != nil {
		return nil, err
	}
	clusterNames := []string{}

	stdOut := out.String()
	lines := strings.Split(stdOut, "\n")
	if len(lines) > 1 {
		for i := range lines {
			tidied := strings.ReplaceAll(lines[i], "*", "")
			tidied = strings.TrimSpace(tidied)
			split := strings.Fields(tidied)
			if len(split) > 0 {
				if strings.Contains(split[0], filter) {
					clusterNames = append(clusterNames, split[0])
				}

			} else if len(tidied) > 0 {
				fmt.Printf("not sure what to do with line: [%v]\n", tidied)
			}
		}
	}

	return clusterNames, nil
}

func SetContext(name string) error {
	contextName = name
	//  engineering@aws-eu-central-1-sandbox
	out, err := runGetStdout("kubectl", "config", "use-context", name)
	if err != nil {
		return err
	}
	if strings.Contains(out.String(), "Switched to context") {
		return nil
	}
	return fmt.Errorf("dont think we switched context:`%v`", out.String())
}
