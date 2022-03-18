package dotnetcoreaspnet_test

import (
	"errors"
	"os"
	"testing"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/dotnet-core-aspnet/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir         string
		buildpackYMLParser *fakes.VersionParser
		detect             packit.DetectFunc
	)

	it.Before(func() {
		workingDir = "some-working-dir"
		buildpackYMLParser = &fakes.VersionParser{}
		detect = dotnetcoreaspnet.Detect(buildpackYMLParser)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("no version requirement", func() {
		it("detects with a plan that provides dotnet-aspnetcore", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
				},
				Provides: []packit.BuildPlanProvision{
					{Name: "dotnet-aspnetcore"},
				},
			}))
		})
	})

	context("when src code contains a buildpack.yml", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "1.2.3"
		})

		it("provides dotnet-aspnetcore and requires specific version of dotnet-aspnetcore", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: "dotnet-aspnetcore",
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
					{
						Name: "dotnet-aspnetcore",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "1.2.3",
						},
					},
				},
			}))
		})
	})

	context("when the BP_DOTNET_FRAMEWORK_VERSION is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_DOTNET_FRAMEWORK_VERSION", "1.2.3")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_DOTNET_FRAMEWORK_VERSION")).To(Succeed())
		})

		it("provides and requires specific version of dotnet-aspnetcore", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: "dotnet-aspnetcore",
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "dotnet-runtime",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
					{
						Name: "dotnet-aspnetcore",
						Metadata: map[string]interface{}{
							"version-source": "BP_DOTNET_FRAMEWORK_VERSION",
							"version":        "1.2.3",
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml"))
			})
		})
	})
}
