package version

import (
	"fmt"
	"runtime"
	rdebug "runtime/debug"
	"strings"
)

var (
	GitCommit           string
	GitBranch           string
	GitSummary          string
	BuildDate           string
	AppVersion          string
	ConditionorcVersion = conditionorcVersion()
	FleetDBVersion      = fleetdbVersion()
	GoVersion           = runtime.Version()
)

type Version struct {
	GitCommit           string `json:"git_commit"`
	GitBranch           string `json:"git_branch"`
	GitSummary          string `json:"git_summary"`
	BuildDate           string `json:"build_date"`
	AppVersion          string `json:"app_version"`
	GoVersion           string `json:"go_version"`
	FleetDBVersion      string `json:"fleetdb_version"`
	ConditionorcVersion string `json:"conditionorc_version"`
}

func Current() *Version {
	return &Version{
		GitBranch:           GitBranch,
		GitCommit:           GitCommit,
		GitSummary:          GitSummary,
		BuildDate:           BuildDate,
		AppVersion:          AppVersion,
		GoVersion:           GoVersion,
		ConditionorcVersion: ConditionorcVersion,
		FleetDBVersion:      FleetDBVersion,
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("version=%s ref=%s branch=%s built=%s", v.AppVersion, v.GitCommit, v.GitBranch, v.BuildDate)
}

func fleetdbVersion() string {
	buildInfo, ok := rdebug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, d := range buildInfo.Deps {
		if strings.Contains(d.Path, "fleetdb") {
			return d.Version
		}
	}

	return ""
}

func conditionorcVersion() string {
	buildInfo, ok := rdebug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, d := range buildInfo.Deps {
		if strings.Contains(d.Path, "conditionorc") {
			return d.Version
		}
	}

	return ""
}
