package dotnetcoreaspnet

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterial(dependency postal.Dependency) packit.BuildpackPlan
}

//go:generate faux --interface Symlinker --output fakes/symlinker.go
type Symlinker interface {
	Link(workingDir, layerPath string) (Err error)
}

func Build(entries EntryResolver, dependencies DependencyManager, planRefinery BuildPlanRefinery, symlinker Symlinker, logger LogEmitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Dotnet Core ASPNet version")

		// if RUNTIME_VERSION env var set,
		// then use it and don't look at the build plan values.
		entry := entries.Resolve(context.Plan.Entries)
		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		aspNetLayer, err := context.Layers.Get("dotnet-core-aspnet")
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := planRefinery.BillOfMaterial(postal.Dependency{
			ID:      dependency.ID,
			Name:    dependency.Name,
			SHA256:  dependency.SHA256,
			Stacks:  dependency.Stacks,
			URI:     dependency.URI,
			Version: dependency.Version,
		})

		cachedSHA, ok := aspNetLayer.Metadata["dependency-sha"].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", aspNetLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{aspNetLayer},
			}, nil
		}
		logger.Process("Executing build process")

		err = aspNetLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		aspNetLayer.Launch = entry.Metadata["launch"] == true
		aspNetLayer.Build = entry.Metadata["build"] == true
		aspNetLayer.Cache = entry.Metadata["build"] == true

		logger.Subprocess("Installing Dotnet Core ASPNet %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, aspNetLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		aspNetLayer.Metadata = map[string]interface{}{
			"dependency-sha": dependency.SHA256,
			"built_at":       clock.Now().Format(time.RFC3339Nano),
		}

		aspNetLayer.SharedEnv.Override("DOTNET_ROOT", filepath.Join(context.WorkingDir, ".dotnet_root"))
		logger.Environment(aspNetLayer.SharedEnv)

		err = symlinker.Link(context.WorkingDir, aspNetLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{aspNetLayer},
		}, nil
	}
}
