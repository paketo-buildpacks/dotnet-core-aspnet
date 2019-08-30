package integration_test

import (
	"path/filepath"
"testing"


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
		Expect func(interface{}, ...interface{}) Assertion
	 	Eventually func(interface{}, ...interface{}) AsyncAssertion
		app    *dagger.App
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
		app, err := dagger.PackBuild(filepath.Join("testdata", "simple_aspnet_app"), runtimeURI, aspnetURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.StartWithCommand("./simple_aspnet_app --server.urls http://0.0.0.0:${PORT}")).To(Succeed())

		Eventually(func() string {
			body, _, _ := app.HTTPGet("/")
			return body
		}).Should(ContainSubstring("Welcome"))

	})

}

