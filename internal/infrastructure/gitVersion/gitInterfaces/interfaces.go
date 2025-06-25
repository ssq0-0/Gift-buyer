package gitInterfaces

import gittypes "gift-buyer/internal/infrastructure/gitVersion/gitTypes"

type GitVersionController interface {
	GetLatestVersion() (*gittypes.GitHubRelease, error)
	GetCurrentVersion() (string, error)
	CompareVersions(localVersion, remoteVersion string) (bool, error)
}
