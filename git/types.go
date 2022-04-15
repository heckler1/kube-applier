package git

// GitUtilInterface allows for mocking out the functionality of GitUtil when testing the full process of an apply run.
type GitUtilInterface interface {
	HeadHash() (string, error)
	ListAllFiles() ([]string, error)
	CommitLog(string) (string, error)
	ListDiffFiles(string, string) ([]string, error)
}

type DiffResult map[string]DiffType

type DiffType string

const (
	DiffTypeModified = "M"
	DiffTypeAdded    = "A"
	DiffTypeRemoved  = "R"
)
