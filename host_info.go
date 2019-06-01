package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const (
	cmdCPU  = "lscpu --json"
	cmdDisk = "lsblk --json --nodeps --bytes"
)

type hostInfo struct {
	cpuModel string
	diskSize int64
}

type lscpuInfo struct {
	Lscpu []struct {
		Field string `json:"field"`
		Data  string `json:"data"`
	} `json:"lscpu"`
}

type lsblkInfo struct {
	Lsblk []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"blockdevices"`
}

func (h *hostInfo) getDiskInfo() error {
	out, err := execShell(cmdDisk)
	if err != nil {
		return err
	}

	i := &lsblkInfo{}
	if err := json.Unmarshal([]byte(out), i); err != nil {
		return err
	}
	size := int64(0)
	for _, b := range i.Lsblk {
		if strings.Contains(b.Name, "loop") {
			continue
		}
		size += b.Size
	}
	h.diskSize = size
	return nil
}

func (h *hostInfo) getCPUInfo() error {
	out, err := execShell(cmdCPU)
	if err != nil {
		return err
	}

	i := &lscpuInfo{}
	if err := json.Unmarshal([]byte(out), i); err != nil {
		return err
	}
	for _, c := range i.Lscpu {
		if strings.Contains(c.Field, "Model name") {
			h.cpuModel = parseCPUModel(c.Data)
			break
		}
	}
	return nil
}

func parseCPUModel(model string) string {
	if strings.Contains(model, "Intel") {
		ms := strings.Split(model, " ")
		return strings.TrimSpace(ms[2])
	}
	return model
}

// ExecBashShell exec bash shell command
func execShell(arg string) (string, error) {
	cmd := exec.Command("bash", "-c", arg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("exec shell failed, out_msg:%s, err_msg:%s, err:%s",
			stdout.String(), stderr.String(), err.Error())
	}
	return stdout.String(), nil
}
