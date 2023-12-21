package frame

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

func updateProcessIdFile(pidFilePath string) (int, error) {
	f, openPidFileErr := os.OpenFile(pidFilePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if openPidFileErr != nil {
		return 0, openPidFileErr
	}
	defer func() {
		_ = f.Close()
	}()

	pidNum := os.Getpid()
	_, writeErr := f.Write([]byte(strconv.Itoa(pidNum)))
	if writeErr != nil {
		return pidNum, writeErr
	}

	return pidNum, nil
}

func checkProcessIdFileExist(pidFilePath string) (bool, error) {
	_, statErr := os.Stat(pidFilePath)
	if statErr == nil {
		return true, nil
	}
	if os.IsNotExist(statErr) {
		return false, nil
	}
	return false, statErr
}

func deleteProcessIdFile(pidFilePath string) error {
	_, statErr := os.Stat(pidFilePath)
	if os.IsNotExist(statErr) {
		return nil
	}

	return os.Remove(pidFilePath)
}
