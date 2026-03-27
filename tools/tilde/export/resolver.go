package export

import (
	"os"
	"path/filepath"
	"strings"
)

// DirResolver resolves URI paths using directory-based index files.
// For a URI like "/post/foo" it looks for /post/foo/index.html (among other candidates).
// Used in development.
type DirResolver struct {
	Root string
}

func (r *DirResolver) Resolve(uriPath string) (string, error) {
	candidates := []string{
		uriPath,
		uriPath + ".html",
		strings.TrimSuffix(uriPath, "/") + "/index.html",
	}
	return resolveFromCandidates(r.Root, candidates)
}

// FlatResolver resolves URI paths using flat .html files.
// For a URI like "/post/foo" it looks for /post/foo.html (among other candidates).
// Used in production.
type FlatResolver struct {
	Root string
}

func (r *FlatResolver) Resolve(uriPath string) (string, error) {
	candidates := []string{
		uriPath,
		uriPath + ".html",
	}
	return resolveFromCandidates(r.Root, candidates)
}

func resolveFromCandidates(root string, candidates []string) (string, error) {
	for _, rel := range candidates {
		abs, ok := safeLocalPath(root, rel)
		if !ok {
			return "", ErrForbidden
		}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return abs, nil
		}
	}
	return "", ErrNotFound
}

// safeLocalPath joins root and relativePath and verifies the result stays within root,
// preventing directory traversal attacks.
func safeLocalPath(root, relativePath string) (string, bool) {
	abs, err := filepath.Abs(filepath.Join(root, relativePath))
	if err != nil {
		return "", false
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	if !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) && abs != rootAbs {
		return "", false
	}
	return abs, true
}
