package build

import "fmt"

// version holds the version of the application
var version = "development"

// commit is the commit id of the input to the build
var commit = "sha1"

// user is who built the app
var user = "build.user"

// time is when the app was built
var time = "build.time"

// VersionInfo holds information about the application's version
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
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
			Version:   version,
			Commit:    commit,
			BuildTime: time,
		},
		AppDescription: "Icon Repository",
	}
}

// GetInfoString constructs and returns a string containing the build info.
func GetInfoString() string {
	return fmt.Sprintf("Version:\t%v\nCommit:\t\t%v\nBuild time:\t%v\nBuild user:\t%v\n", version, commit, time, user)
}
