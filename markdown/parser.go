package markdown

import (
	"bytes"
	"fmt"
	ext "markdown-live-preview/extension"
	"markdown-live-preview/resource"
	"os"
	"path/filepath"

	"github.com/gohugoio/hugo-goldmark-extensions/extras"
	attributes "github.com/mdigger/goldmark-attributes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

const TEMPLATE = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<base href="%s">
<style>%s</style>
<style>%s</style>
<style>%s</style>
</head>
<body>%s</body>
<script>%s</script>
<script>%s</script>
</html>`

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.CJK,
		extension.DefinitionList,
		extension.Footnote,
		extension.GFM,
		extension.Table,
		extension.TaskList,
		extras.New(extras.Config{
			Delete:      extras.DeleteConfig{Enable: true},
			Insert:      extras.InsertConfig{Enable: true},
			Mark:        extras.MarkConfig{Enable: true},
			Subscript:   extras.SubscriptConfig{Enable: true},
			Superscript: extras.SuperscriptConfig{Enable: true},
		}),
		// my extensions
		ext.NewHSGalleryImageExtension(),
		ext.NewHSImgExtension(),
		ext.NewImageExtension(),
		&frontmatter.Extender{},
	),
	goldmark.WithParserOptions(
		parser.WithAttribute(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithUnsafe(),
	),
	attributes.Enable,
)

func fillTemplate(path string, content string) string {
	dir := filepath.ToSlash(filepath.Dir(path)) + "/"
	return fmt.Sprintf(TEMPLATE, dir, resource.FileDarkCSS, resource.FileGLightboxCSS, resource.FileStyleCSS, content, resource.FileGLightboxJS, resource.FileScriptJS)
}

func RenderMarkdown(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return fillTemplate(path, "Failed to read file"), err
	}

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return fillTemplate(path, "Failed to parse file"), err
	}

	// Post Processing
	result := ext.PostProcessHSImg(buf)

	return fillTemplate(path, result), nil
}
