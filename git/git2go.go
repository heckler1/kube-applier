//go:build !git_cli
// +build !git_cli

package git

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/box/kube-applier/applylist"
	git "github.com/libgit2/git2go/v33"
)

type LibGitUtil struct {
	repo *git.Repository
}

func NewGitUtil(repoPath string) (GitUtilInterface, error) {
	util := &LibGitUtil{}
	var err error
	util.repo, err = git.OpenRepository(repoPath)
	return util, err
}

// HeadHash returns the hash of the current HEAD commit.
func (g *LibGitUtil) HeadHash() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", err
	}
	commit, err := g.repo.LookupCommit(head.Target())

	return commit.Id().String(), err
}

// CommitLog returns the log of the specified commit, including a list of the files that were modified.
func (g *LibGitUtil) CommitLog(hash string) (string, error) {
	commit, err := g.repo.LookupCommit(git.NewOidFromBytes([]byte(hash)))

	logMsgTempl := template.New(`commit {{ .commitId }}
	Author: {{ .author }}
	Date:   {{ .date }}

			{{ .message }}

	{{ .modified }}`)

	logMsgTempl.Option("missingkey=error")

	commitData := map[string]string{}
	commitData["commitId"] = commit.Id().String()
	commitData["author"] = commit.Author().Name + " <" + commit.Author().Email + ">"
	commitData["date"] = commit.Author().When.Format(time.UnixDate)
	commitData["message"] = commit.Message()

	commitTree, err := commit.Tree()
	previousTree, err := commit.Parent(0).Tree()

	diffResult := make(DiffResult)

	internalDiff, err := g.repo.DiffTreeToTree(previousTree, commitTree, &git.DiffOptions{})
	if err != nil {
		return "", err
	}

	nDelta, err := internalDiff.NumDeltas()
	if err != nil {
		return "", err
	}

	for i := 0; i < nDelta; i++ {
		delta, err := internalDiff.Delta(i)
		if err != nil {
			return "", err
		}
		switch delta.Status {
		case git.DeltaUnmodified:
			continue
		case git.DeltaAdded, git.DeltaUntracked:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
		case git.DeltaDeleted:
			diffResult[delta.OldFile.Path] = DiffTypeRemoved
		case git.DeltaCopied:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
		case git.DeltaRenamed:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
			diffResult[delta.OldFile.Path] = DiffTypeRemoved
		case git.DeltaModified, git.DeltaTypeChange:
			if delta.OldFile.Path != "" {
				diffResult[delta.OldFile.Path] = DiffTypeModified
			}
			if delta.NewFile.Path != "" {
				diffResult[delta.NewFile.Path] = DiffTypeModified
			}
		default:
			return "", fmt.Errorf("unhandled diff type %s", delta.Status)
		}

	}

	diffLines := []string{}
	for path, difftype := range diffResult {
		diffLines = append(diffLines, fmt.Sprintf("%s\t%s", difftype, path))
	}

	commitData["modified"] = strings.Join(diffLines, "\n")

	builder := &strings.Builder{}
	err = logMsgTempl.Execute(builder, commitData)

	return builder.String(), err
}

// ListAllFiles returns a list of all files under $REPO_PATH, with paths relative to $REPO_PATH.
func (g *LibGitUtil) ListAllFiles() ([]string, error) {

	idx, err := g.repo.Index()
	if err != nil {
		return nil, err
	}

	files := []string{}
	for i := uint(0); i < idx.EntryCount(); i++ {
		entry, err := idx.EntryByIndex(i)
		if err != nil {
			return nil, err
		}

		files = append(files, entry.Path)

		fmt.Printf("entry.Path: %v\n", entry.Path)
	}

	// Are they relative or full?
	fullPaths := applylist.PrependToEachPath(g.repo.Path(), files)
	return fullPaths, nil
}

// ListDiffFiles returns the file names that were added, modified, copied, or renamed.
// Deletes are ignored because kube-applier should not apply files deleted by a commit.
func (g *LibGitUtil) ListDiffFiles(oldHash, newHash string) ([]string, error) {

	fromOid, err := git.NewOid(oldHash)
	if err != nil {
		return nil, err
	}
	toOid, err := git.NewOid(newHash)
	if err != nil {
		return nil, err
	}

	fromCommit, err := g.repo.LookupCommit(fromOid)
	if err != nil {
		return nil, err
	}
	toCommit, err := g.repo.LookupCommit(toOid)
	if err != nil {
		return nil, err
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, err
	}
	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, err
	}
	diffOpts := &git.DiffOptions{}
	//if len(limitToFiles) > 0 {
	//	diffOpts.Pathspec = limitToFiles
	//}

	diffResult := make(DiffResult)

	internalDiff, err := g.repo.DiffTreeToTree(fromTree, toTree, diffOpts)
	if err != nil {
		return nil, err
	}

	nDelta, err := internalDiff.NumDeltas()
	if err != nil {
		return nil, err
	}

	for i := 0; i < nDelta; i++ {
		delta, err := internalDiff.Delta(i)
		if err != nil {
			return nil, err
		}
		switch delta.Status {
		case git.DeltaUnmodified:
			continue
		case git.DeltaAdded, git.DeltaUntracked:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
		case git.DeltaDeleted:
			continue
		case git.DeltaCopied:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
		case git.DeltaRenamed:
			diffResult[delta.NewFile.Path] = DiffTypeAdded
			diffResult[delta.OldFile.Path] = DiffTypeRemoved
		case git.DeltaModified, git.DeltaTypeChange:
			if delta.OldFile.Path != "" {
				diffResult[delta.OldFile.Path] = DiffTypeModified
			}
			if delta.NewFile.Path != "" {
				diffResult[delta.NewFile.Path] = DiffTypeModified
			}
		default:
			return nil, fmt.Errorf("unhandled diff type %s", delta.Status)
		}
	}

	relativePaths := []string{}
	for path := range diffResult {
		relativePaths = append(relativePaths, path)
	}

	fullPaths := applylist.PrependToEachPath(g.repo.Path(), relativePaths)
	return fullPaths, nil
}
