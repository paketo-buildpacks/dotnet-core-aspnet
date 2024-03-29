package main

import (
	"os"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	buildpackYMLParser := dotnetcoreaspnet.NewBuildpackYMLParser()
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	dotnetRootLinker := dotnetcoreaspnet.NewDotnetRootLinker()

	packit.Run(
		dotnetcoreaspnet.Detect(buildpackYMLParser),
		dotnetcoreaspnet.Build(
			entryResolver,
			dependencyManager,
			dotnetRootLinker,
			Generator{},
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
