package integration_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	bpDir, aspnetURI, runtimeURI string
)

var suite = spec.New("Integration", spec.Report(report.Terminal{}))

func init() {
	suite("Integration", testIntegration)
}

func TestIntegration(t *testing.T) {
	var err error
	Expect := NewWithT(t).Expect
	bpDir, err = dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())
	aspnetURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(aspnetURI)

	runtimeURI, err = dagger.GetLatestBuildpack("dotnet-core-runtime-cnb")
	Expect(err).ToNot(HaveOccurred())
	defer dagger.DeleteBuildpack(runtimeURI)

	suite.Run(t)
}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect     func(interface{}, ...interface{}) Assertion
		Eventually func(interface{}, ...interface{}) AsyncAssertion
		app        *dagger.App
		err        error
	)

	it.Before(func() {
		Expect = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
	})

	it.After(func() {
		if app != nil {
			app.Destroy()
		}
	})

	it("should build a working OCI image for a simple app with aspnet dependencies", func() {
		app, err = dagger.NewPack(
			filepath.Join("testdata", "simple_aspnet_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("./simple_aspnet_app --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Welcome"))

	})

	it("should build a working OCI image for a simple app with aspnet dependencies that has a buildpack.yml in it", func() {
		majorMinor := "2.2"
		version, err := getLowestRuntimeVersionInMajorMinor(majorMinor)
		Expect(err).ToNot(HaveOccurred())
		bpYml := fmt.Sprintf(`---
dotnet-framework:
  version: "%s"
`, version)

		bpYmlPath := filepath.Join("testdata", "simple_aspnet_app_with_buildpack_yml", "buildpack.yml")
		Expect(ioutil.WriteFile(bpYmlPath, []byte(bpYml), 0644)).To(Succeed())

		app, err = dagger.NewPack(
			filepath.Join("testdata", "simple_aspnet_app_with_buildpack_yml"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(runtimeURI, aspnetURI),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("./simple_aspnet_app --urls http://0.0.0.0:${PORT}")).To(Succeed())

		Expect(app.BuildLogs()).To(ContainSubstring(fmt.Sprintf("dotnet-runtime.%s", version)))
		Expect(app.BuildLogs()).To(ContainSubstring(fmt.Sprintf("dotnet-aspnetcore.%s", version)))

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("simple_aspnet_app"))

	})

}

func getLowestRuntimeVersionInMajorMinor(majorMinor string) (string, error) {
	type buildpackTomlVersion struct {
		Metadata struct {
			Dependencies []struct {
				Version string `toml:"version"`
			} `toml:"dependencies"`
		} `toml:"metadata"`
	}

	bpToml := buildpackTomlVersion{}
	output, err := ioutil.ReadFile(filepath.Join("..", "buildpack.toml"))
	if err != nil {
		return "", err
	}

	majorMinorConstraint, err := semver.NewConstraint(fmt.Sprintf("%s.*", majorMinor))
	if err != nil {
		return "", err
	}

	lowestVersion, err := semver.NewVersion(fmt.Sprintf("%s.99", majorMinor))
	if err != nil {
		return "", err
	}

	_, err = toml.Decode(string(output), &bpToml)
	if err != nil {
		return "", err
	}

	for _, dep := range bpToml.Metadata.Dependencies {
		depVersion, err := semver.NewVersion(dep.Version)
		if err != nil {
			return "", err
		}
		if majorMinorConstraint.Check(depVersion) {
			if depVersion.LessThan(lowestVersion) {
				lowestVersion = depVersion
			}
		}
	}

	return lowestVersion.String(), nil
}
