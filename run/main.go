package main

import (
	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	buildpackYMLParser := dotnetcoreaspnet.NewBuildpackYMLParser()
	packit.Run(
		dotnetcoreaspnet.Detect(buildpackYMLParser),
		dotnetcoreaspnet.Build(),
	)
}
