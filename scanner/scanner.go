package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type RepoContext struct {
	TotalFiles   int
	MaxDepth     int
	Extensions   map[string]int
	RootFiles    []string
	AllFilenames []string
	FolderNames  []string
	DocContent   string
}

func ScanRepo(root string) (RepoContext, error) {
	ctx := RepoContext{
		Extensions: make(map[string]int),
	}

	rootEntries, _ := os.ReadDir(root)
	for _, e := range rootEntries {
		if !e.IsDir() {
			ctx.RootFiles = append(ctx.RootFiles, e.Name())
		}
	}

	ctx.DocContent = readDocs(root)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip hidden dirs (like .venv, .agents, etc)
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// track folder names
		if d.IsDir() && path != root {
			ctx.FolderNames = append(ctx.FolderNames, d.Name())
			depth := strings.Count(strings.TrimPrefix(path, root), string(os.PathSeparator))
			if depth > ctx.MaxDepth {
				ctx.MaxDepth = depth
			}
			return nil
		}

		if !d.IsDir() {
			ctx.TotalFiles++
			ctx.AllFilenames = append(ctx.AllFilenames, d.Name())
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext != "" {
				ctx.Extensions[ext]++
			}
		}

		return nil
	})

	return ctx, err
}

// look for docs for better context
func readDocs(root string) string {
	candidates := []string{
		"README.md", "README.txt", "README",
		"CONTRIBUTING.md", "ARCHITECTURE.md", "DESIGN.md",
	}

	var collected strings.Builder

	for _, name := range candidates {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err == nil {
			collected.WriteString("-- " + name + " --\n")
			// truncate to 500 chars per file to avoid huge LLM prompts
			content := string(data)
			if len(content) > 500 {
				content = content[:500] + "...(truncated)"
			}
			collected.WriteString(content + "\n\n")
		}
	}

	docsDir := filepath.Join(root, "docs")
	entries, err := os.ReadDir(docsDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				data, err := os.ReadFile(filepath.Join(docsDir, e.Name()))
				if err == nil {
					collected.WriteString("-- docs/" + e.Name() + " --\n")
					content := string(data)
					if len(content) > 300 {
						content = content[:300] + "...(truncated)"
					}
					collected.WriteString(content + "\n\n")
				}
			}
		}
	}

	return collected.String()
}
