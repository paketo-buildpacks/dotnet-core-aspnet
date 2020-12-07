package main

import (
	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	packit.Run(dotnetcoreaspnet.Detect(), dotnetcoreaspnet.Build())
}
