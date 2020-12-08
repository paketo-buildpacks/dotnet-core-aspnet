package dotnetcoreaspnet_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/paketo-buildpacks/dotnet-core-aspnet/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		workingDir        string
		cnbDir            string
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		clock             chronos.Clock
		timeStamp         time.Time
		planRefinery      *fakes.BuildPlanRefinery
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "dotnet-aspnet",
			Metadata: map[string]interface{}{
				"version-source": "buildpack.yml",
				"version":        "2.5.x",
				"launch":         true,
			},
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:   "dotnet-aspnet",
			Name: "Dotnet Core ASPNet",
		}

		planRefinery = &fakes.BuildPlanRefinery{}

		planRefinery.BillOfMaterialCall.Returns.BuildpackPlan = packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{
				{
					Name: "dotnet-aspnet",
					Metadata: map[string]interface{}{
						"version-source": "buildpack.yml",
						"version":        "2.5.x",
						"launch":         true,
					},
				},
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := dotnetcoreaspnet.NewLogEmitter(buffer)

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		build = dotnetcoreaspnet.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "dotnet-aspnet",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "2.5.x",
							"launch":         true,
						},
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "dotnet-aspnet",
						Metadata: map[string]interface{}{
							"version-source": "buildpack.yml",
							"version":        "2.5.x",
							"launch":         true,
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name: "dotnet-core-aspnet",
					Path: filepath.Join(layersDir, "dotnet-core-aspnet"),
					SharedEnv: packit.Environment{
						"DOTNET_ROOT.override": filepath.Join(workingDir, ".dotnet_root"),
					},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    true,
					Cache:     false,
					Metadata: map[string]interface{}{
						"dependency-sha": "",
						"built_at":       timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))
		Expect(filepath.Join(workingDir, ".dotnet_root")).To(BeADirectory())
	})
}
