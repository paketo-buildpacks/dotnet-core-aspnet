package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-aspnet-cnb/aspnet"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

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

	runtimeConfig, err := aspnet.NewRuntimeConfig(context.Application.Root)
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
