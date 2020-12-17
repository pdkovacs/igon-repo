package endpoints

import "github.com/pdkovacs/igo-repo/backend/pkg/config"

// VersionInfo holds information about the application's version
type VersionInfo struct {
	// Version is the application version
	Version string `json:"version"`
	// Commit is the commit id of source code which was the input to build
	Commit string `json:"commit"`
}

// Info holds basic information about the application
type Info struct {
	VersionInfo    VersionInfo `json:"versionInfo"`
	AppDescription string      `json:"appDescription"`
}

// GetInfo returns basic information about the application
func GetInfo() Info {
	return Info{
		VersionInfo: VersionInfo{
			Version: config.Version,
			Commit:  config.Commit,
		},
		AppDescription: "Icon Repository",
	}
}
