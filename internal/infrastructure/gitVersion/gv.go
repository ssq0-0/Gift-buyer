package gitVersion

import (
	"encoding/json"
	"fmt"
	"gift-buyer/internal/infrastructure/gitVersion/gitInterfaces"
	gittypes "gift-buyer/internal/infrastructure/gitVersion/gitTypes"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitVersionControllerImpl struct {
	owner    string
	repoName string
	apiLink  string
}

func NewGitVersionController(owner, repoName, apiLink string) gitInterfaces.GitVersionController {
	return &GitVersionControllerImpl{
		owner:    owner,
		repoName: repoName,
		apiLink:  apiLink,
	}
}

func (gvc *GitVersionControllerImpl) GetLatestVersion() (*gittypes.GitHubRelease, error) {
	return gvc.getLatestGitHubRelease()
}

func (gvc *GitVersionControllerImpl) GetCurrentVersion() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}

	return gvc.getLatestLocalTag(repo)
}

func (gvc *GitVersionControllerImpl) CompareVersions(localVersion, remoteVersion string) (bool, error) {
	if localVersion == "" || remoteVersion == "" {
		return false, fmt.Errorf("local or remote version is empty")
	}

	local, err := semver.NewVersion(localVersion)
	if err != nil {
		return false, err
	}

	remoteVersion = strings.TrimPrefix(remoteVersion, "v")
	remote, err := semver.NewVersion(remoteVersion)
	if err != nil {
		return false, err
	}

	return remote.GreaterThan(local), nil
}

func (gvc *GitVersionControllerImpl) getLatestLocalTag(repo *git.Repository) (string, error) {
	refIter, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var versions []*semver.Version

	refIter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() {
			tagName := ref.Name().Short()
			v, err := semver.NewVersion(strings.TrimPrefix(tagName, "v"))
			if err == nil {
				versions = append(versions, v)
			}
		}
		return nil
	})

	if len(versions) == 0 {
		return "", fmt.Errorf("no valid tags found")
	}

	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions[0].String(), nil
}

func (gvc *GitVersionControllerImpl) getLatestGitHubRelease() (*gittypes.GitHubRelease, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", gvc.owner, gvc.repoName))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release gittypes.GitHubRelease
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}
