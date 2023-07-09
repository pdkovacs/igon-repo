package gitrepo

import (
	"fmt"
	"iconrepo/internal/app/domain"
	"path/filepath"
)

type iconfilePathComponents struct {
	pathToFormatDir      string
	pathToSizeDir        string
	pathToIconfile       string
	pathToIconfileInRepo string
}

type filePaths struct {
	pathPrefix string
}

func NewGitFilePaths(pathPrefix string) filePaths {
	return filePaths{pathPrefix: pathPrefix}
}

func getFileName(iconName string, format string, size string) string {
	return fmt.Sprintf("%s@%s.%s", iconName, size, format)
}

func (p filePaths) getPathComponents0(iconName string, format string, size string) iconfilePathComponents {
	fileName := getFileName(iconName, format, size)
	pathToFormatDir := filepath.Join(p.pathPrefix, format)
	pathToSizeDir := filepath.Join(pathToFormatDir, size)
	pathToIconfile := filepath.Join(pathToSizeDir, fileName)
	pathToIconfileInRepo := filepath.Join(format, filepath.Join(size, fileName))
	return iconfilePathComponents{
		pathToFormatDir,
		pathToSizeDir,
		pathToIconfile,
		pathToIconfileInRepo,
	}
}

func (p filePaths) getPathComponents(iconName string, iconfile domain.IconfileDescriptor) iconfilePathComponents {
	return p.getPathComponents0(
		iconName,
		iconfile.Format,
		iconfile.Size,
	)
}

func (p filePaths) GetAbsolutePathToIconfile(iconName string, iconfile domain.IconfileDescriptor) string {
	return p.getPathComponents(iconName, iconfile).pathToIconfile
}

func (p filePaths) GetPathToIconfileInRepo(iconName string, iconfile domain.IconfileDescriptor) string {
	return p.getPathComponents(iconName, iconfile).pathToIconfileInRepo
}
