package dotnetcoreaspnet_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetCoreAspnet(t *testing.T) {
	suite := spec.New("dotnet-core-aspnet", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("DotnetRootLinker", testDotnetRootLinker)
	suite.Run(t)
}
