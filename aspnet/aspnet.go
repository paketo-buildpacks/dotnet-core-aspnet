package aspnet

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"os"
	"path/filepath"
)

const DotnetAspNet = "dotnet-aspnet"

type Contributor struct {
	context      build.Build
	plan         buildpackplan.Plan
	aspnetLayer layers.DependencyLayer
	aspnetRuntimeLayer layers.Layer
	logger       logger.Logger
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantDependency, err := context.Plans.GetShallowMerged(DotnetAspNet)
	if err != nil{
		return Contributor{}, false, err
	}
	if !wantDependency {
		return Contributor{}, false, nil
	}

	dep, err := context.Buildpack.RuntimeDependency(DotnetAspNet, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}


	return Contributor{
		context:      context,
		plan:         plan,
		aspnetLayer: context.Layers.DependencyLayer(dep),
		aspnetRuntimeLayer: context.Layers.Layer("aspnetRuntime"),
		logger:       context.Logger,
	}, true, nil
}

func (c Contributor) Contribute() error {
	err := c.aspnetLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.Body("Expanding to %s", layer.Root)

		if err := helper.ExtractTarXz(artifact, layer.Root, 0); err != nil {
			return err
		}

		return nil
	}, getFlags(c.plan.Metadata)...)

	if err != nil{
		return err
	}

	err = c.aspnetRuntimeLayer.Contribute(c.context.Buildpack, func(layer layers.Layer) error {
		pathToRuntime := os.Getenv("DOTNET_ROOT")
		runtimeFiles, err := filepath.Glob(filepath.Join(pathToRuntime, "shared", "*"))
		for _, file := range runtimeFiles {
			if err := helper.WriteSymlink(file, filepath.Join(layer.Root, "shared", filepath.Base(file))); err != nil {
				return err
			}
		}

		aspnetFiles, err := filepath.Glob(filepath.Join(c.aspnetLayer.Root, "shared", "*"))
		if err != nil {
			return err
		}
		for _, file := range aspnetFiles {
			if err := helper.WriteSymlink(file, filepath.Join(layer.Root, "shared", filepath.Base(file))); err != nil {
				return err
			}
		}

		hostDir := filepath.Join(pathToRuntime, "host")

		if err := helper.WriteSymlink(hostDir, filepath.Join(layer.Root, filepath.Base(hostDir))); err != nil{
			return err
		}

		if err := layer.OverrideSharedEnv("DOTNET_ROOT", filepath.Join(layer.Root)); err != nil {
			return err
		}

		return nil
	}, getFlags(c.plan.Metadata)...)

	if err != nil{
		return err
	}

	return nil
}

func getFlags(metadata buildpackplan.Metadata) []layers.Flag{
	flagsArray := []layers.Flag{}
	flagValueMap := map[string]layers.Flag {"build": layers.Build, "launch": layers.Launch, "cache": layers.Cache}
	for _, flagName := range []string{"build", "launch", "cache"} {
		flagPresent, _ := metadata[flagName].(bool)
		if flagPresent {
			flagsArray = append(flagsArray, flagValueMap[flagName])
		}
	}
	return flagsArray
}