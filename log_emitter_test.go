package dotnetcoreaspnet_test

import (
	"bytes"
	"testing"
	"time"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter dotnetcoreaspnet.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = dotnetcoreaspnet.NewLogEmitter(buffer)
	})

	context("SelectedDependency", func() {
		it("prints details about the selected dependency", func() {
			entry := packit.BuildpackPlanEntry{
				Metadata: map[string]interface{}{
					"version-source": "some-source",
				},
			}
			dependency := postal.Dependency{
				Name:    "Dotnet Core ASPNet",
				Version: "some-version",
			}

			emitter.SelectedDependency(entry, dependency, time.Now())
			Expect(buffer.String()).To(Equal("    Selected Dotnet Core ASPNet version (using some-source): some-version\n\n"))
		})

		context("when the version source is missing", func() {
			it("prints details about the selected dependency", func() {
				dependency := postal.Dependency{
					Name:    "Dotnet Core ASPNet",
					Version: "some-version",
				}

				emitter.SelectedDependency(packit.BuildpackPlanEntry{}, dependency, time.Now())
				Expect(buffer.String()).To(Equal("    Selected Dotnet Core ASPNet version (using <unknown>): some-version\n\n"))
			})
		})

		context("when it is within 30 days of the deprecation date", func() {
			it("returns a warning that the dependency will be deprecated after the deprecation date", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(-29 * 24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Dotnet Core ASPNet",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Dotnet Core ASPNet version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Dotnet Core ASPNet will be deprecated after 2021-04-01.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Dotnet Core ASPNet before this time.\n\n"))
			})
		})

		context("when it is on the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Dotnet Core ASPNet",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Dotnet Core ASPNet version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Dotnet Core ASPNet is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Dotnet Core ASPNet.\n\n"))
			})
		})

		context("when it is after the the deprecation date", func() {
			it("returns a warning that the version of the dependency is no longer supported", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				now := deprecationDate.Add(24 * time.Hour)

				entry := packit.BuildpackPlanEntry{
					Metadata: map[string]interface{}{"version-source": "some-source"},
				}
				dependency := postal.Dependency{
					DeprecationDate: deprecationDate,
					Name:            "Dotnet Core ASPNet",
					Version:         "some-version",
				}

				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Dotnet Core ASPNet version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Dotnet Core ASPNet is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Dotnet Core ASPNet.\n\n"))
			})
		})
	})

	context("Candidates", func() {
		it("prints a formatted map of version source inputs", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnet",
					Metadata: map[string]interface{}{
						"version-source": "package.json",
						"version":        "package-json-version",
					},
				},
				{
					Name: "dotnet-aspnet",
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
				{
					Name: "dotnet-aspnet",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "buildpack-yml-version",
					},
				},
				{
					Name: "dotnet-aspnet",
				},
			})

			Expect(buffer.String()).To(ContainSubstring("    Candidate version sources (in priority order):"))
			Expect(buffer.String()).To(ContainSubstring("      buildpack.yml -> \"buildpack-yml-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"other-version\""))
			Expect(buffer.String()).To(ContainSubstring("      <unknown>     -> \"*\""))
		})
	})

	context("Environment", func() {
		it("prints details about the environment", func() {
			emitter.Environment(packit.Environment{
				"GEM_PATH.override": "/some/path",
			})

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    GEM_PATH -> \"/some/path\""))
		})
	})
}
