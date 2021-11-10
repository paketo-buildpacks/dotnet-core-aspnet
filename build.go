package dotnetcoreaspnet

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface Symlinker --output fakes/symlinker.go
type Symlinker interface {
	Link(workingDir, layerPath string) (Err error)
}

func Build(entries EntryResolver, dependencies DependencyManager, symlinker Symlinker, logger LogEmitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Dotnet Core ASPNet version")

		if v, ok := os.LookupEnv("RUNTIME_VERSION"); ok {
			context.Plan.Entries = append(context.Plan.Entries, packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version":        v,
					"version-source": "RUNTIME_VERSION",
				},
			})
		}

		priorities := []interface{}{
			"RUNTIME_VERSION",
			"BP_DOTNET_FRAMEWORK_VERSION",
			"buildpack.yml",
			regexp.MustCompile(`.*\.(cs)|(fs)|(vb)proj`),
			"runtimeconfig.json",
		}

		entry, sortedEntries := entries.Resolve("dotnet-aspnetcore", context.Plan.Entries, priorities)
		logger.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the .NET Framework version through buildpack.yml will be deprecated soon in Dotnet Core ASPNet Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_DOTNET_FRAMEWORK_VERSION environment variable instead. See docs for more information.")
			logger.Break()
		}

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		aspNetLayer, err := context.Layers.Get("dotnet-core-aspnet")
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entries.MergeLayerTypes("dotnet-aspnetcore", context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		cachedSHA, ok := aspNetLayer.Metadata["dependency-sha"].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", aspNetLayer.Path)
			logger.Break()

			err = symlinker.Link(context.WorkingDir, aspNetLayer.Path)
			if err != nil {
				return packit.BuildResult{}, err
			}

			aspNetLayer.Launch, aspNetLayer.Build, aspNetLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{aspNetLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}
		logger.Process("Executing build process")

		aspNetLayer, err = aspNetLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		aspNetLayer.Launch, aspNetLayer.Build, aspNetLayer.Cache = launch, build, launch || build

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
			Layers: []packit.Layer{aspNetLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
