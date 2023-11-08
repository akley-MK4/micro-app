package frame

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func startSubprocess(execPath string, kwArgs map[string]interface{}) (*exec.Cmd, string, error) {
	var arg []string
	kwArgs["process_type"] = SubProcessType
	for k, v := range kwArgs {
		arg = append(arg, fmt.Sprintf("-%s=%v", k, v))
	}

	cmd := exec.Command(execPath, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	//_, outputErr := cmd.Output()
	//if outputErr != nil {
	//	logger.WarnFmt("startSubprocess, %v", outputErr)
	//}

	argsStr := fmt.Sprintf("%s %s", execPath, strings.Join(arg, " "))
	if err := cmd.Start(); err != nil {
		return nil, argsStr, err
	}

	return cmd, argsStr, nil
}
