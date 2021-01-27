package dotnetcoreaspnet_test

import (
	"bytes"
	"testing"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer   *bytes.Buffer
		resolver dotnetcoreaspnet.PlanEntryResolver
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		resolver = dotnetcoreaspnet.NewPlanEntryResolver(dotnetcoreaspnet.NewLogEmitter(buffer))
	})

	context("when a buildpack.yml entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "buildpack-yml-version",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
		})
	})

	context("when buildpack.yml and project file entries are included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "my-app.fsproj",
						"version":        "project-file-version",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "buildpack.yml",
					"version":        "buildpack-yml-version",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      my-app.fsproj -> \"project-file-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
		})
	})

	context("when buildpack.yml and RUNTIME_VERSION entries are included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "RUNTIME_VERSION",
						"version":        "runtime-version",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "RUNTIME_VERSION",
					"version":        "runtime-version",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      RUNTIME_VERSION -> \"runtime-version\""))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml   -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>       -> \"other-version\""))
		})
	})

	context("when project file and unknown entries are included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version":        "other-version",
						"version-source": "unknown source",
					},
				},
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "my-app.csproj",
						"version":        "project-file-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version-source": "my-app.csproj",
					"version":        "project-file-version",
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      my-app.csproj  -> \"project-file-version\""))
			Expect(buffer.String()).To(ContainSubstring("      unknown source -> \"other-version\""))
		})
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name: "dotnet-aspnetcore",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "buildpack-yml-version",
						},
					},
					{
						Name: "dotnet-aspnetcore",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
						"build":          true,
					},
				}))
			})
		})
	})

	context("when an unknown source entry is included", func() {
		it("resolves the best plan entry", func() {
			entry := resolver.Resolve([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnetcore",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "dotnet-aspnetcore",
				Metadata: map[string]interface{}{
					"version": "other-version",
				},
			}))
		})
	})
}
