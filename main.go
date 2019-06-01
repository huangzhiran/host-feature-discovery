// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	// Bin : Name of the binary
	Bin = "host-feature-discovery"
)

var (
	// Version : Version of the binary
	// This will be set using ldflags at compile time
	Version = ""
)

func main() {

	log.SetPrefix(Bin + ": ")

	if Version == "" {
		log.Print("Version is not set.")
		log.Fatal("Be sure to compile with '-ldflags \"-X main.Version=${HFD_VERSION}\"' and to set $HFD_VERSION")
	}

	log.Printf("Running %s in version %s", Bin, Version)

	conf := Conf{}
	conf.getConfFromArgv(os.Args)
	conf.getConfFromEnv()
	log.Print("Loaded configuration:")
	log.Print("Oneshot: ", conf.Oneshot)
	log.Print("SleepInterval: ", conf.SleepInterval)
	log.Print("OutputFilePath: ", conf.OutputFilePath)

	log.Print("Start running")
	if err := run(conf); err != nil {
		log.Printf("Unexpected error: %v", err)
	}
	log.Print("Exiting")
}

func run(conf Conf) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	exitChan := make(chan bool)

	go func() {
		select {
		case s := <-sigChan:
			log.Printf("Received signal \"%v\", shutting down.", s)
			exitChan <- true
		}
	}()

	outputFileAbsPath, err := filepath.Abs(conf.OutputFilePath)
	if err != nil {
		return fmt.Errorf("Failed to retrieve absolute path of output file: %v", err)
	}
	tmpDirPath := filepath.Dir(outputFileAbsPath) + "/hfd-tmp"

	err = os.Mkdir(tmpDirPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("Failed to create temporary directory: %v", err)
	}

L:
	for {
		tmpOutputFile, err := ioutil.TempFile(tmpDirPath, "hfd-")
		if err != nil {
			return fmt.Errorf("Fail to create temporary output file: %v", err)
		}

		info := hostInfo{}
		if err := info.getCPUInfo(); err != nil {
			return err
		}
		if err := info.getDiskInfo(); err != nil {
			return err
		}

		log.Print("Writing labels to output file")
		fmt.Fprintf(tmpOutputFile, "cpu-model=%s\n", info.cpuModel)
		fmt.Fprintf(tmpOutputFile, "disk-size-bytes=%d\n", info.diskSize)

		err = tmpOutputFile.Chmod(0644)
		if err != nil {
			return fmt.Errorf("Error chmod temporary file: %v", err)
		}

		err = tmpOutputFile.Close()
		if err != nil {
			return fmt.Errorf("Error closing temporary file: %v", err)
		}

		err = os.Rename(tmpOutputFile.Name(), conf.OutputFilePath)
		if err != nil {
			return fmt.Errorf("Error moving temporary file '%s': %v", conf.OutputFilePath, err)
		}

		if conf.Oneshot {
			break
		}

		log.Print("Sleeping for ", conf.SleepInterval)

		select {
		case <-exitChan:
			break L
		case <-time.After(conf.SleepInterval):
			break
		}
	}

	return nil
}
