package main

import (
	"os"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYMLParser := dotnetcoreaspnet.NewBuildpackYMLParser()
	logEmitter := dotnetcoreaspnet.NewLogEmitter(os.Stdout)
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	dotnetRootLinker := dotnetcoreaspnet.NewDotnetRootLinker()

	packit.Run(
		dotnetcoreaspnet.Detect(buildpackYMLParser),
		dotnetcoreaspnet.Build(
			entryResolver,
			dependencyManager,
			dotnetRootLinker,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
