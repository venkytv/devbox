// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package docker

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/imdario/mergo"
)

// This package provides an API to build images using docker.
// The API is implemented by calling the docker CLI directly.
//
// This actually might be a better approach than relying on the docker go libraries:
// + Those libraries require the Docker client to be installed anyways, so they don't
//   save the user from having to install Docker as a dependency.
// + The libraries are pretty bloated, increasing the size of our binaries.
// + In the past we've had some trouble with docker go libraries, and dependency
//   management.

type BuildFlags struct {
	Name           string
	Tags           []string
	Platforms      []string
	NoCache        bool
	DockerfilePath string
}

type BuildOptions func(*BuildFlags)

func WithFlags(src *BuildFlags) BuildOptions {
	return func(dst *BuildFlags) {
		err := mergo.Merge(dst, src, mergo.WithOverride)
		if err != nil {
			panic(err)
		}
	}
}

func WithoutCache() BuildOptions {
	return func(flags *BuildFlags) {
		flags.NoCache = true
	}
}

func Build(path string, opts ...BuildOptions) error {
	flags := &BuildFlags{}
	for _, opt := range opts {
		opt(flags)
	}

	args := []string{"build", "."}
	args = ToArgs(args, flags)

	dir, fileName := parsePath(path)
	if fileName != "" {
		args = append(args, "-f", fileName)
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "BUILDKIT=1")
	cmd.Dir = dir
	return cmd.Run()
}

func parsePath(path string) (string, string) {
	// If the path points to a file that exists, separate the directory part
	// and the file part:
	if isFile(path) {
		return filepath.Dir(path), filepath.Base(path)
	} else {
		// Otherwise assume the entire thing is a directory:
		return path, ""
	}
}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.Mode().IsRegular()
}
