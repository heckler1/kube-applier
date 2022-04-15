//go:build git_cli
// +build git_cli

package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/box/kube-applier/applylist"
)

// GitUtil allows for fetching information about a Git repository using Git CLI commands.
type GitCLIUtil struct {
	RepoPath string
}

func NewGitUtil(repoPath string) (GitUtilInterface, error) {
	return &git.GitCLIUtil{RepoPath: repoPath}
}

// HeadHash returns the hash of the current HEAD commit.
func (g *GitCLIUtil) HeadHash() (string, error) {
	hash, err := runGitCmd(g.RepoPath, "rev-parse", "HEAD")
	return strings.TrimSuffix(hash, "\n"), err
}

// CommitLog returns the log of the specified commit, including a list of the files that were modified.
func (g *GitCLIUtil) CommitLog(hash string) (string, error) {
	log, err := runGitCmd(g.RepoPath, "log", "-1", "--name-status", hash)
	return log, err
}

// ListAllFiles returns a list of all files under $REPO_PATH, with paths relative to $REPO_PATH.
func (g *GitCLIUtil) ListAllFiles() ([]string, error) {
	raw, err := runGitCmd(g.RepoPath, "ls-files")
	if err != nil {
		return nil, err
	}
	relativePaths := strings.Split(raw, "\n")
	fullPaths := applylist.PrependToEachPath(g.RepoPath, relativePaths)
	return fullPaths, nil
}

// ListDiffFiles returns the file names that were added, modified, copied, or renamed.
// Deletes are ignored because kube-applier should not apply files deleted by a commit.
func (g *GitCLIUtil) ListDiffFiles(oldHash, newHash string) ([]string, error) {
	raw, err := runGitCmd(g.RepoPath, "diff", "--diff-filter=AMCR", "--name-only", "--relative", oldHash, newHash)
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return []string{}, nil
	}
	relativePaths := strings.Split(raw, "\n")
	fullPaths := applylist.PrependToEachPath(g.RepoPath, relativePaths)
	return fullPaths, nil
}

func runGitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command %v: %v: %s", strings.Join(cmd.Args, " "), err, output)
	}
	return string(output), nil
}
