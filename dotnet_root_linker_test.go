package dotnetcoreaspnet_test

import (
	"os"
	"path/filepath"
	"testing"

	dotnetcoreaspnet "github.com/paketo-buildpacks/dotnet-core-aspnet"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDotnetRootLinker(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		dotnetLinker dotnetcoreaspnet.DotnetRootLinker
		workingDir   string
		layerPath    string
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		layerPath, err = os.MkdirTemp("", "layer-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(layerPath, "shared", "dir1"), os.ModePerm)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(layerPath, "shared", "dir2"), os.ModePerm)).To(Succeed())

		dotnetLinker = dotnetcoreaspnet.NewDotnetRootLinker()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		Expect(os.RemoveAll(layerPath)).To(Succeed())
	})

	context("Link", func() {
		it("creates a .dotnet_root dir in workspace with symlink to layerpath", func() {
			err := dotnetLinker.Link(workingDir, layerPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(workingDir, ".dotnet_root")).To(BeADirectory())

			fi, err := os.Lstat(filepath.Join(workingDir, ".dotnet_root", "shared", "dir1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Mode() & os.ModeSymlink).ToNot(BeZero())

			link, err := os.Readlink(filepath.Join(workingDir, ".dotnet_root", "shared", "dir1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal(filepath.Join(layerPath, "shared", "dir1")))

			fi, err = os.Lstat(filepath.Join(workingDir, ".dotnet_root", "shared", "dir2"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Mode() & os.ModeSymlink).ToNot(BeZero())

			link, err = os.Readlink(filepath.Join(workingDir, ".dotnet_root", "shared", "dir2"))
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal(filepath.Join(layerPath, "shared", "dir2")))
		})

		context("error cases", func() {
			context("when the '.dotnet_root' dir can not be created", func() {
				it.Before(func() {
					Expect(os.Chmod(filepath.Join(workingDir), 0000)).To(Succeed())
				})
				it("errors", func() {
					err := dotnetLinker.Link(workingDir, layerPath)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when there is a bad file glob", func() {
				it("returns an error", func() {
					err := dotnetLinker.Link(workingDir, `\`)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("syntax error in pattern")))
				})
			})

			context("when the symlink can not be created", func() {
				it.Before(func() {
					Expect(os.MkdirAll(filepath.Join(workingDir, ".dotnet_root", "shared", "dir1"), os.ModePerm)).To(Succeed())
				})

				it("errors", func() {
					err := dotnetLinker.Link(workingDir, layerPath)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("file exists")))
				})
			})
		})
	})
}
