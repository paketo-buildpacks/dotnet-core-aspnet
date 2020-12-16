package main

import (
	"os"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYMLParser := dotnetcoreaspnet.NewBuildpackYMLParser()
	logEmitter := dotnetcoreaspnet.NewLogEmitter(os.Stdout)
	entryResolver := dotnetcoreaspnet.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := dotnetcoreaspnet.NewPlanRefinery()
	dotnetRootLinker := dotnetcoreaspnet.NewDotnetRootLinker()

	packit.Run(
		dotnetcoreaspnet.Detect(buildpackYMLParser),
		dotnetcoreaspnet.Build(
			entryResolver,
			dependencyManager,
			planRefinery,
			dotnetRootLinker,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
