package jinja

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderer_RejectsFilesystemTemplateLoading(t *testing.T) {
	t.Parallel()

	const secret = "TOP_SECRET_HOST_FILE_CONTENTS"

	dir := t.TempDir()
	absPath := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(absPath, []byte(secret), 0o600))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	traversalPath, err := filepath.Rel(cwd, absPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(traversalPath, ".."), "expected a ../ traversal path, got %q", traversalPath)

	paths := map[string]string{"absolute": absPath, "traversal": traversalPath}
	tags := func(path string) map[string]string {
		return map[string]string{
			"include": `{% include "` + path + `" %}`,
			"import":  `{% import "` + path + `" as m %}`,
			"from":    `{% from "` + path + `" import thing %}`,
			"extends": `{% extends "` + path + `" %}`,
		}
	}

	for pathName, path := range paths {
		for tag, tmpl := range tags(path) {
			t.Run(pathName+"/"+tag, func(t *testing.T) {
				t.Parallel()
				out, err := NewRenderer(Context{}).Render(tmpl)
				require.Error(t, err)
				require.Equal(t, errTemplateFileAccessDenied.Error(), err.Error())
				require.NotContains(t, out, secret)
			})
		}
	}
}

func TestRenderer_SandboxKeepsNormalTemplatesWorking(t *testing.T) {
	t.Parallel()

	out, err := NewRenderer(Context{}).Render(`{% for i in [1, 2, 3] %}{{ i }}{% endfor %}`)
	require.NoError(t, err)
	require.Equal(t, "123", out)
}
