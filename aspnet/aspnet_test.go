package aspnet

import (
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUnitDotnet(t *testing.T) {
	spec.Run(t, "Detect", testDotnet, spec.Report(report.Terminal{}))
}

func testDotnet(t *testing.T, when spec.G, it spec.S) {
	var (
		factory     *test.BuildFactory
		stubDotnetAspnetFixture = filepath.Join("testdata", "stub-dotnet-aspnet.tar.xz")
		symlinkPath string

	)

	it.Before(func() {
		var err error

		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		factory.AddDependency(DotnetAspNet, stubDotnetAspnetFixture)

		symlinkPath, err = ioutil.TempDir(os.TempDir(), "runtime")
		Expect(err).ToNot(HaveOccurred())

		os.Setenv("DOTNET_ROOT", symlinkPath)
	})

	it.After(func () {
		os.RemoveAll(symlinkPath)
		os.Unsetenv("DOTNET_ROOT")
	})

	when("runtime.NewContributor", func() {
		it("returns true if a build plan exists", func() {
			factory.AddPlan(buildpackplan.Plan{Name: DotnetAspNet})

			_, willContribute, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			_, willContribute, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("Contribute", func() {
		it("installs the aspnet dependency", func() {
			factory.AddPlan(buildpackplan.Plan{Name: DotnetAspNet})

			dotnetASPNetContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetASPNetContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(DotnetAspNet)
			Expect(filepath.Join(layer.Root, "stub-dir", "stub.txt")).To(BeARegularFile())
		})

		it("uses the default version when a version is not requested", func() {
			factory.AddDependencyWithVersion(DotnetAspNet, "0.9", filepath.Join("testdata", "stub-dotnet-aspnet-default.tar.xz"))
			factory.SetDefaultVersion(DotnetAspNet, "0.9")
			factory.AddPlan(buildpackplan.Plan{Name: DotnetAspNet})

			dotnetASPNetContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetASPNetContributor.Contribute()).To(Succeed())
			layer := factory.Build.Layers.Layer(DotnetAspNet)
			Expect(layer).To(test.HaveLayerVersion("0.9"))
		})

		it("contributes dotnet runtime to the build layer when included in the build plan", func() {
			factory.AddPlan(buildpackplan.Plan{
				Name: DotnetAspNet,
				Metadata: buildpackplan.Metadata{
					"build": true,
				},
			})

			dotnetASPNetContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetASPNetContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(DotnetAspNet)
			Expect(layer).To(test.HaveLayerMetadata(true, false, false))
		})

		it("contributes dotnet runtime to the launch layer when included in the build plan", func() {
			factory.AddPlan(buildpackplan.Plan{
				Name: DotnetAspNet,
				Metadata: buildpackplan.Metadata{
					"launch": true,
				},
			})

			dotnetASPNetContributor, _, err := NewContributor(factory.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(dotnetASPNetContributor.Contribute()).To(Succeed())

			layer := factory.Build.Layers.Layer(DotnetAspNet)
			Expect(layer).To(test.HaveLayerMetadata(false, false, true))
		})

		it("returns an error when unsupported version of dotnet runtime is included in the build plan", func() {
			factory.AddPlan(buildpackplan.Plan{
				Name:    DotnetAspNet,
				Version: "9000.0.0",
				Metadata: buildpackplan.Metadata{
					"launch": true,
				},
			})

			_, shouldContribute, err := NewContributor(factory.Build)
			Expect(err).To(HaveOccurred())
			Expect(shouldContribute).To(BeFalse())
		})
	})
}
