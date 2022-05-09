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
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("an app is rebuilt and aspnet dependency is unchanged", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err             error
				logs            fmt.Stringer
				firstImage      occam.Image
				secondImage     occam.Image
				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					dotnetCoreRuntimeBuildpack.Online,
					buildpack,
					buildPlanBuildpack,
				)

			firstImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(3))

			Expect(firstImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))

			firstContainer, err = docker.Container.Run.
				WithCommand(
					fmt.Sprintf(`test -f /layers/%s/dotnet-core-aspnet/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					test -f /workspace/.dotnet_root/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					echo 'AspNetCore.dll exists' &&
					sleep infinity`,
						strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(firstContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("AspNetCore.dll exists"))

			// Second pack build
			secondImage, logs, err = build.Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(3))

			Expect(secondImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).NotTo(ContainSubstring("  Executing build process"))
			Expect(logs.String()).To(ContainSubstring(fmt.Sprintf("  Reusing cached layer /layers/%s/dotnet-core-aspnet", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))))

			Expect(secondImage.Buildpacks[1].Layers["dotnet-core-aspnet"].SHA).To(Equal(firstImage.Buildpacks[1].Layers["dotnet-core-aspnet"].SHA))

			secondContainer, err = docker.Container.Run.
				WithCommand(
					fmt.Sprintf(`test -f /layers/%s/dotnet-core-aspnet/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					test -f /workspace/.dotnet_root/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					echo 'AspNetCore.dll exists' &&
					sleep infinity`,
						strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("AspNetCore.dll exists"))
		})
	})

	context("an app is rebuilt and requirement changes", func() {
		it("does not reuse a layer from the previous build", func() {
			var (
				err             error
				logs            fmt.Stringer
				firstImage      occam.Image
				secondImage     occam.Image
				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					dotnetCoreRuntimeBuildpack.Online,
					buildpack,
					buildPlanBuildpack,
				)

			firstImage, logs, err = build.WithEnv(map[string]string{
				"BP_DOTNET_FRAMEWORK_VERSION": "3.*",
			}).Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(3))

			Expect(firstImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))

			firstContainer, err = docker.Container.Run.
				WithCommand(
					fmt.Sprintf(`test -f /layers/%s/dotnet-core-aspnet/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					test -f /workspace/.dotnet_root/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					echo 'AspNetCore.dll exists' &&
					sleep infinity`,
						strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(firstContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("AspNetCore.dll exists"))

			// Second pack build
			secondImage, logs, err = build.WithEnv(map[string]string{
				"BP_DOTNET_FRAMEWORK_VERSION": "6.*",
			}).Execute(name, source)

			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(3))

			Expect(secondImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("dotnet-core-aspnet"))

			Expect(logs.String()).To(ContainSubstring("  Executing build process"))
			Expect(logs.String()).NotTo(ContainSubstring("Reusing cached layer"))

			Expect(secondImage.Buildpacks[1].Layers["dotnet-core-aspnet"].SHA).NotTo(Equal(firstImage.Buildpacks[1].Layers["dotnet-core-aspnet"].SHA))

			secondContainer, err = docker.Container.Run.
				WithCommand(
					fmt.Sprintf(`test -f /layers/%s/dotnet-core-aspnet/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					test -f /workspace/.dotnet_root/shared/Microsoft.AspNetCore.App/*/Microsoft.AspNetCore.dll &&
					echo 'AspNetCore.dll exists' &&
					sleep infinity`,
						strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring("AspNetCore.dll exists"))
		})
	})
}
