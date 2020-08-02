package version

import (
	"bytes"
	"fmt"
)

var (
	// The commit ID, filled in by the build process
	GitCommit string

	// The main version number
	Version = "0.4.0"

	// Optional suffix -- empty string for releases, "dev" for development, or
	// a beta, rc, etc. indicator
	VersionPrerelease = "dev"
)

type VersionInfo struct {
	CommitId          string
	Version           string
	VersionPrerelease string
}

func GetVersion() *VersionInfo {
	return &VersionInfo{
		CommitId:          GitCommit,
		Version:           Version,
		VersionPrerelease: VersionPrerelease,
	}
}

func (v *VersionInfo) String() string {
	version := fmt.Sprintf("%s", v.Version)

	if v.VersionPrerelease != "" {
		version = fmt.Sprintf("%s-%s", version, v.VersionPrerelease)
	}

	return version
}

func (v *VersionInfo) FullVersionNumber(includeCommitId bool) string {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "zone-update v%s", v.Version)
	if v.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", v.VersionPrerelease)
	}

	if includeCommitId && v.CommitId != "" {
		fmt.Fprintf(&versionString, " (%s)", v.CommitId)
	}

	return versionString.String()
}