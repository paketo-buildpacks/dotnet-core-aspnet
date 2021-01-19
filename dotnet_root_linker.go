package dotnetcoreaspnet

import (
	"os"
	"path/filepath"
)

type DotnetRootLinker struct{}

func NewDotnetRootLinker() DotnetRootLinker {
	return DotnetRootLinker{}
}

func (dl DotnetRootLinker) Link(workingDir, layerPath string) error {
	err := os.MkdirAll(filepath.Join(workingDir, ".dotnet_root", "shared"), os.ModePerm)
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(layerPath, "shared", "*"))
	if err != nil {
		return err
	}

	for _, f := range files {
		filename := filepath.Base(f)
		err := os.Symlink(filepath.Join(layerPath, "shared", filename), filepath.Join(workingDir, ".dotnet_root", "shared", filename))
		if err != nil {
			return err
		}
	}

	return nil
}
