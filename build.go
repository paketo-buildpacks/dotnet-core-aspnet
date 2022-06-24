package dotnetcoreaspnet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/Masterminds/semver"
	"github.com/gravityblast/go-jsmin"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface Symlinker --output fakes/symlinker.go
type Symlinker interface {
	Link(workingDir, layerPath string) (Err error)
}

//go:generate faux --interface VersionResolver --output fakes/version_resolver.go
type VersionResolver interface {
	Resolve(path string, entry packit.BuildpackPlanEntry, stack string) (postal.Dependency, error)
}

func Build(entries EntryResolver, versionResolver VersionResolver, dependencies DependencyManager, symlinker Symlinker, logger LogEmitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving Dotnet Core ASPNet version")

		runtimeV, aspnetV, err := configParse(filepath.Join(context.WorkingDir, "*.runtimeconfig.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return packit.BuildResult{}, err
		}

		if aspnetV != "" {
			context.Plan.Entries = append(context.Plan.Entries, packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version":        aspnetV,
					"version-source": "runtimeconfig.json ASP.NET",
				},
			})
		}

		if runtimeV != "" {
			context.Plan.Entries = append(context.Plan.Entries, packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version":        runtimeV,
					"version-source": "runtimeconfig.json .NET Runtime",
				},
			})
		}

		priorities := []interface{}{
			"BP_DOTNET_FRAMEWORK_VERSION",
			"buildpack.yml",
			regexp.MustCompile(`.*\.(cs)|(fs)|(vb)proj`),
			"runtimeconfig.json ASP.NET",
			"runtimeconfig.json .NET Runtime",
			"runtimeconfig.json",
			".NET Execute Buildpack",
		}

		entry, sortedEntries := entries.Resolve("dotnet-aspnetcore", context.Plan.Entries, priorities)
		logger.Candidates(sortedEntries)

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the .NET Framework version through buildpack.yml will be deprecated soon in Dotnet Core ASPNet Buildpack v%s.", nextMajorVersion.String())
			logger.Subprocess("Please specify the version through the $BP_DOTNET_FRAMEWORK_VERSION environment variable instead. See docs for more information.")
			logger.Break()
		}

		if source == ".NET Execute Buildpack" {
			logger.Subprocess("No version of ASP.NET requested. Skipping install.")

			return packit.BuildResult{}, nil
		}

		dependency, err := versionResolver.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry, context.Stack)
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
			return dependencies.Deliver(dependency, context.CNBPath, aspNetLayer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		aspNetLayer.Metadata = map[string]interface{}{
			"dependency-sha": dependency.SHA256,
		}

		aspNetLayer.SharedEnv.Override("DOTNET_ROOT", aspNetLayer.Path)
		logger.Environment(aspNetLayer.SharedEnv)

		return packit.BuildResult{
			Layers: []packit.Layer{aspNetLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}

type RuntimeConfig struct {
	RuntimeVersion string
	ASPNETVersion  string
}

type framework struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func configParse(glob string) (runtime, aspnet string, err error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return "", "", fmt.Errorf("failed to find *.runtimeconfig.json: %w: %q", err, glob)
	}

	if len(files) > 1 {
		return "", "", fmt.Errorf("multiple *.runtimeconfig.json files present: %v", files)
	}

	if len(files) == 0 {
		return "", "", fmt.Errorf("no *.runtimeconfig.json found: %w", os.ErrNotExist)
	}

	var data struct {
		RuntimeOptions struct {
			Framework  framework   `json:"framework"`
			Frameworks []framework `json:"frameworks"`
		} `json:"runtimeOptions"`
	}

	file, err := os.Open(files[0])
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	buffer := bytes.NewBuffer(nil)
	err = jsmin.Min(file, buffer)
	if err != nil {
		return "", "", err
	}

	err = json.NewDecoder(buffer).Decode(&data)
	if err != nil {
		return "", "", err
	}

	switch data.RuntimeOptions.Framework.Name {
	case "Microsoft.NETCore.App":
		runtime = versionOrWildcard(data.RuntimeOptions.Framework.Version)
	case "Microsoft.AspNetCore.App":
		aspnet = versionOrWildcard(data.RuntimeOptions.Framework.Version)
		runtime = aspnet
	default:
		runtime = ""
		aspnet = ""
	}

	for _, f := range data.RuntimeOptions.Frameworks {
		switch f.Name {
		case "Microsoft.NETCore.App":
			if runtime != "" {
				return "", "", fmt.Errorf("malformed runtimeconfig.json: multiple '%s' frameworks specified", f.Name)
			}
			runtime = versionOrWildcard(f.Version)
		case "Microsoft.AspNetCore.App":
			if aspnet != "" {
				return "", "", fmt.Errorf("malformed runtimeconfig.json: multiple '%s' frameworks specified", f.Name)
			}
			aspnet = versionOrWildcard(f.Version)
		default:
			continue
		}
	}
	return runtime, aspnet, nil
}

func versionOrWildcard(version string) string {
	if version == "" {
		return "*"
	}
	return version
}
