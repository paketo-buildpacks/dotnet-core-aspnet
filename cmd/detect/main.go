package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-aspnet-cnb/aspnet"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
)

type BuildpackYAML struct {
	Config struct {
		Version string `yaml:"version""`
	} `yaml:"dotnet-aspnet"`
}

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	plan := buildplan.Plan{
		Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}}}

	runtimeConfig, err := utils.NewRuntimeConfig(context.Application.Root)
	if err != nil {
		return context.Fail(), err
	}

	hasFDE, err := runtimeConfig.HasExecutable()
	if err != nil {
		return context.Fail(), err
	}

	if runtimeConfig.HasASPNetDependency() {
		if hasFDE {

			plan.Requires = []buildplan.Required{{
				Name:     aspnet.DotnetAspNet,
				Version:  runtimeConfig.Version,
				Metadata: buildplan.Metadata{"launch": true},
			}, {
				Name:     "dotnet-runtime",
				Version:  runtimeConfig.Version,
				Metadata: buildplan.Metadata{"build": true, "launch": true},
			}}
		}
	}

	return context.Pass(plan)
}

func checkIfVersionsAreValid(versionRuntimeConfig, versionBuildpackYAML string) error {
	splitVersionRuntimeConfig := strings.Split(versionRuntimeConfig, ".")
	splitVersionBuildpackYAML := strings.Split(versionBuildpackYAML, ".")

	if splitVersionBuildpackYAML[0] != splitVersionRuntimeConfig[0] {
		return fmt.Errorf("major versions of runtimes do not match between buildpack.yml and runtimeconfig.json")
	}

	minorBPYAML, err := strconv.Atoi(splitVersionBuildpackYAML[1])
	if err != nil {
		return err
	}

	minorRuntimeConfig, err := strconv.Atoi(splitVersionRuntimeConfig[1])
	if err != nil {
		return err
	}

	if minorBPYAML < minorRuntimeConfig {
		return fmt.Errorf("the minor version of the runtimeconfig.json is greater than the minor version of the buildpack.yml")
	}

	return nil
}

func rollForward(version string, context detect.Detect) (string, bool, error) {
	splitVersion := strings.Split(version, ".")
	anyPatch := fmt.Sprintf("%s.%s.*", splitVersion[0], splitVersion[1])
	anyMinor := fmt.Sprintf("%s.*.*", splitVersion[0])

	versions := []string{version, anyPatch, anyMinor}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return "", false, err
	}

	for _, versionConstraint := range versions {
		highestVersion, err := deps.Best(aspnet.DotnetAspNet, versionConstraint, context.Stack)
		if err == nil {
			return highestVersion.Version.Original(), true, nil
		}
	}

	return "", false, fmt.Errorf("no compatible versions found")
}

func LoadBuildpackYAML(appRoot string) (BuildpackYAML, error) {
	var err error
	buildpackYAML := BuildpackYAML{}
	bpYamlPath := filepath.Join(appRoot, "buildpack.yml")

	if exists, err := helper.FileExists(bpYamlPath); err != nil {
		return BuildpackYAML{}, err
	} else if exists {
		err = helper.ReadBuildpackYaml(bpYamlPath, &buildpackYAML)
	}
	return buildpackYAML, err
}
