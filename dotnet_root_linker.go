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

	err = os.Symlink(filepath.Join(layerPath, "shared", "Microsoft.AspNetCore.App"), filepath.Join(workingDir, ".dotnet_root", "shared", "Microsoft.AspNetCore.App"))
	if err != nil {
		return err
	}

	return nil
}
