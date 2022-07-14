package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a container with dotnet aspnet", func() {
		var (
			image     occam.Image
			container occam.Container

			name    string
			source  string
			sbomDir string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			sbomDir, err = os.MkdirTemp("", "sbom")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())

			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(os.RemoveAll(sbomDir)).To(Succeed())
		})

		it("builds an oci image with aspnet dlls installed", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					dotnetCoreRuntimeBuildpack.Online,
					buildpack,
					buildPlanBuildpack,
				).
				WithSBOMOutputDir(sbomDir).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving .NET Core ASPNet version",
				"    Candidate version sources (in priority order):",
				MatchRegexp(`      RUNTIME_VERSION -> "\d+\.\d+\.\d+"`),
				"      <unknown>       -> \"\"",
				"",
				MatchRegexp(`    Selected .NET Core ASPNet version \(using RUNTIME_VERSION\): \d+\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing .NET Core ASPNet \d+\.\d+\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring launch environment",
				`    DOTNET_ROOT -> "/workspace/.dotnet_root"`,
			))

			container, err = docker.Container.Run.
				WithCommand(
					fmt.Sprintf(`test -f /layers/%s/dotnet-core-aspnet/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					test -f /workspace/.dotnet_root/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					echo 'AspNetCore.dll exists' &&
					sleep infinity`,
						strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("AspNetCore.dll exists"))

			contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", "sbom.legacy.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`"name":".NET Core ASPNet"`))

			// check that all required SBOM files are present
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "dotnet-core-aspnet", "sbom.cdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "dotnet-core-aspnet", "sbom.spdx.json")).To(BeARegularFile())
			Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "dotnet-core-aspnet", "sbom.syft.json")).To(BeARegularFile())

			// check an SBOM file
			contents, err = os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "dotnet-core-aspnet", "sbom.cdx.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring(`"name": ".NET Core ASPNet"`))
		})
	})

	context("image is built with BP_DOTNET_FRAMEWORK_VERSION set", func() {
		var (
			image  occam.Image
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("considers BP_DOTNET_FRAMEWORK_VERSION as a version source", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					dotnetCoreRuntimeBuildpack.Online,
					buildpack,
					buildPlanBuildpack,
				).
				WithEnv(map[string]string{"BP_DOTNET_FRAMEWORK_VERSION": "3.1.*"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving .NET Core ASPNet version",
				"    Candidate version sources (in priority order):",
				MatchRegexp(`      RUNTIME_VERSION             -> "3\.1\.\d+"`),
				"      BP_DOTNET_FRAMEWORK_VERSION -> \"3.1.*\"",
				"      <unknown>                   -> \"\"",
				"",
				MatchRegexp(`    Selected .NET Core ASPNet version \(using RUNTIME_VERSION\): \d+\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing .NET Core ASPNet 3\.1\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				"",
				"  Configuring launch environment",
				`    DOTNET_ROOT -> "/workspace/.dotnet_root"`,
			))
		})
	}, spec.Sequential())
}
