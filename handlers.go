package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/bep/golibsass/libsass"
	"github.com/gosimple/slug"
	"github.com/tdewolff/minify/v2"
	cssmin "github.com/tdewolff/minify/v2/css"
	htmlmin "github.com/tdewolff/minify/v2/html"
	"golang.org/x/net/html"
)

// Contains all filepaths already handled with their new path for HTML use.
var processedFiles map[string]string

// Minifier
var min = minify.New()

// SASS transpiler
var sassTranspiler libsass.Transpiler

func init() {
	min.AddFunc("text/html", htmlmin.Minify)
	min.AddFunc("text/css", cssmin.Minify)

	// Prepare SASS transpiler
	var err error
	if sassTranspiler, err = libsass.New(libsass.Options{}); err != nil {
		log.Panic(err)
	}
}

// processSource processes a source reference (`index.html` an subsequent `href`, `src`, etc.).
// If `src` references a file in the project, the file is processed
// and its final path in outDir is returned with the original querystring or fragment (if any).
func processSource(src string) string {
	src = strings.TrimSpace(src)

	u, err := url.Parse(src)
	if err != nil {
		logError("Invalid URL encountered: ", src)
		return src
	}

	// If src is absolute, it's not a project file reference: return as is
	if u.IsAbs() || filepath.IsAbs(u.Path) {
		return u.String()
	}

	// If file has already been processed, return the complete URL with its new path
	if newPath, ok := processedFiles[u.Path]; ok {
		u.Path = newPath
		return u.String()
	}

	// Open file
	f, err := os.Open(u.Path)
	if err != nil {
		if os.IsNotExist(err) {
			logError(u.Path, " referenced but not found in project")
			return u.String()
		}
		logError("Cannot open "+u.Path+": ", err)
		return u.String()
	}
	defer f.Close()

	// Process file according to its extension
	b := new(bytes.Buffer)
	ext := filepath.Ext(u.Path)
	switch ext {
	case ".html":
		return handleHTML(u.Path)
	case ".css":
		if err = processCSS(b, f); err != nil {
			logFatal("Cannot process "+u.Path+":", err)
		}
	case ".sass", ".scss":
		if err = processSass(b, f); err != nil {
			logFatal("Cannot process "+u.Path+":", err)
		}
	default:
		// For other file types, just copy
		if _, err = io.Copy(b, f); err != nil {
			logFatal("Cannot copy "+u.Path+": ", err)
			return u.String()
		}
	}

	// Set new filename
	u.Path = fmt.Sprintf("/%s.%s%s",
		slug.Make(strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path))),
		smallHash(bytes.NewReader(b.Bytes())),
		ext,
	)

	// Write file to outDir
	if err = os.WriteFile(filepath.Join(outDir, u.Path), b.Bytes(), 0644); err != nil {
		logFatal(err)
	}

	// Store the new path in processedFiles
	processedFiles[f.Name()] = u.Path
	return u.String()
}

// HTML

func processHTML(w io.Writer, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return min.Minify("text/html", w, strings.NewReader(string(b)))
}

func handleHTML(fp string) string {
	finalFilepath := filepath.Join("/",
		strings.TrimSuffix(
			strings.TrimSuffix(fp, "index.html"),
			".html"))

	// Standardize path
	// FIXME: Use filepath.Clean until finally returning a path
	finalPath := path.Clean(path.Join(filepath.SplitList(finalFilepath)...))

	// Done must be added before handling other *.html files to avoid infinite loop on cyclic references
	processedFiles[fp] = finalPath

	// Parse template
	tmpl, data := parseWithLayout(fp, nil)

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		logFatal("Cannot execute ", fp, ": ", err)
	}

	// Handle other files in generated template
	doc, err := html.Parse(&b)
	if err != nil {
		panic(err)
	}
	walkNode(doc)

	// Create output file
	if err = os.MkdirAll(filepath.Join(outDir, finalFilepath), 0755); err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(outDir, finalFilepath, "index.html"))
	if err != nil {
		panic(err)
	}

	// Minify HTML and write to file
	b.Reset() // Reuse buffer
	if err = html.Render(&b, doc); err != nil {
		panic(err)
	}
	if err = min.Minify("text/html", f, &b); err != nil {
		panic(err)
	}

	return finalPath
}

// Handle frontmatter and returns the HTML template embedded in its layout(s).
func parseWithLayout(fp string, data map[string]interface{}) (*template.Template, map[string]interface{}) {
	f, err := os.Open(fp)
	if err != nil {
		if os.IsNotExist(err) {
			logError(fp, " doesn't exist")
			return template.New(""), data
		}
		panic(err)
	}
	defer f.Close()

	var front map[string]interface{}
	body, err := frontmatter.Parse(f, &front)
	if err != nil {
		panic(err)
	}

	// Add frontmatter as template data
	if data == nil {
		data = make(map[string]interface{})
	}
	for k, v := range front {
		data[k] = v
	}

	page := template.New("").Funcs(templateFuncs())
	if layout, _ := front["layout"].(string); layout != "" {
		page, data = parseWithLayout(layout, data)
	}
	page = template.Must(page.Parse(string(body)))
	return page, data
}

// walkNode searches for file references in HTML node and replaces them with final filepaths of build.
func walkNode(n *html.Node) {
	// Replace file references
	if n.Type == html.ElementNode {
		for i, v := range n.Attr {
			if v.Key == "href" || v.Key == "src" {
				n.Attr[i].Val = processSource(v.Val)
			}
		}
	}

	// Recurse on children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNode(c)
	}
}

// CSS

func processCSS(w io.Writer, r io.Reader) error {
	if err := min.Minify("text/css", w, r); err != nil {
		return fmt.Errorf("cannot minify: %w", err)
	}
	return nil
}

// SASS

func processSass(w io.Writer, r io.Reader) error {
	// Read input content for transpilation
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	// Transpile
	res, err := sassTranspiler.Execute(string(b))
	if err != nil {
		return fmt.Errorf("cannot transpile: %w", err)
	}

	// Minify
	if err := min.Minify("text/css", w, strings.NewReader(res.CSS)); err != nil {
		return fmt.Errorf("cannot minify: %w", err)
	}
	return nil
}
