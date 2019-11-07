package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dotnet-core-aspnet-cnb/aspnet"
	"github.com/cloudfoundry/dotnet-core-runtime-cnb/runtime"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
		fakeBuildpackToml := `
[[dependencies]]
id = "dotnet-aspnetcore"
name = "Dotnet ASPNet"
stacks = ["org.testeroni"]
uri = "some-uri"
version = "2.2.5"
`
		_, err := toml.Decode(fakeBuildpackToml, &factory.Detect.Buildpack.Metadata)
		Expect(err).ToNot(HaveOccurred())
		factory.Detect.Stack = "org.testeroni"

	})

	when("the app has a FDE", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(factory.Detect.Application.Root, "appName"), []byte(`fake exe`), os.ModePerm)).To(Succeed())
		})

		it("passes when there is a valid runtimeconfig.json where the specified version of Microsoft.AspNetCore.App is provided", func() {
			runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.AspNetCore.App",
      "version": "2.2.5"
    }
  }
}
`), os.ModePerm)).To(Succeed())
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
				Requires: []buildplan.Required{{
					Name:     aspnet.DotnetAspNet,
					Version:  "2.2.5",
					Metadata: buildplan.Metadata{"launch": true},
				}, {
					Name:     runtime.DotnetRuntime,
					Version:  "2.2.5",
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				}},
			}))
		})

		it("passes when there is a valid runtimeconfig.json where the specified version of Microsoft.AspNetCore.All is provided", func() {
			runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.AspNetCore.All",
      "version": "2.2.5"
    }
  }
}
`), os.ModePerm)).To(Succeed())
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
				Requires: []buildplan.Required{{
					Name:     aspnet.DotnetAspNet,
					Version:  "2.2.5",
					Metadata: buildplan.Metadata{"launch": true},
				}, {
					Name:     runtime.DotnetRuntime,
					Version:  "2.2.5",
					Metadata: buildplan.Metadata{"build": true, "launch": true},
				}},
			}))
		})

		it("passes with no require when there is a valid runtimeconfig.json where there is not a specified version Microsoft.NETCore.App/All", func() {
			runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
  "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.NETCore.App",
      "version": "2.2.5"
    }
  }
}
`), os.ModePerm)).To(Succeed())
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
			}))
		})

		it("passes when there is a valid runtimeconfig.json where there are no runtime options meaning the app is a self contained deployment", func() {
			runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
"runtimeOptions": {}
}
`), os.ModePerm)).To(Succeed())
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
			}))
		})

		it("passes when there is no valid runtimeconfig.json meaning that app is source based", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
			}))
		})
	})

	when("the app does not have a FDE", func() {
		it("passes when there is a valid runtimeconfig.json", func() {
			runtimeConfigJSONPath := filepath.Join(factory.Detect.Application.Root, "appName.runtimeconfig.json")
			Expect(ioutil.WriteFile(runtimeConfigJSONPath, []byte(`
{
   "runtimeOptions": {
    "tfm": "netcoreapp2.2",
    "framework": {
      "name": "Microsoft.AspNetCore.App",
      "version": "1.1.0"
    }
  }
}
`), os.ModePerm)).To(Succeed())
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Plans.Plan).To(Equal(buildplan.Plan{
				Provides: []buildplan.Provided{{Name: aspnet.DotnetAspNet}},
			}))
		})
	})
}
