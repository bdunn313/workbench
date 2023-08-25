// Package embedpkg provides access to the embedded scripts.
package embedpkg

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed scripts/*
var ScriptsFS embed.FS

func Scripts() embed.FS {
	return ScriptsFS
}

type Script struct {
	Name      string
	Filename  string
	Extension string
	Path      string
}

func GetScripts() ([]Script, error) {
	scriptsFS := Scripts()
	entries, err := fs.ReadDir(scriptsFS, "scripts")
	if err != nil {
		return nil, err
	}

	var scripts []Script
	for _, entry := range entries {
		if !entry.IsDir() {
			nameWithoutExtension := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			scripts = append(scripts, Script{
				Name:      nameWithoutExtension,
				Filename:  entry.Name(),
				Extension: filepath.Ext(entry.Name()),
				Path:      filepath.Join("scripts", entry.Name()),
			})
		}
	}
	return scripts, nil
}
