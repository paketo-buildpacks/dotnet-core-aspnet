package dotnetcoreaspnet_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetCoreAspnet(t *testing.T) {
	suite := spec.New("dotnet-core-aspnet", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite.Run(t)
}
