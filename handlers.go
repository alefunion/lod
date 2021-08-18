package main

import (
	"bytes"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/bep/golibsass/libsass"
	"github.com/evanw/esbuild/pkg/api"
	es "github.com/evanw/esbuild/pkg/api"
	"github.com/tdewolff/minify/v2"
	htmlmin "github.com/tdewolff/minify/v2/html"
	"golang.org/x/net/html"
)

// Contains all filepaths already handled with their new path for HTML use.
var done map[string]string

// Minifier
var min = minify.New()

func init() {
	min.AddFunc("text/html", htmlmin.Minify)
}

// fp is the filepath relative to working directory.
// Returns the final path inside HTML build.
func handleFile(fp string) string {
	if newfp, ok := done[fp]; ok {
		return newfp
	}
	if !isRealtiveFile(fp) {
		return fp
	}
	if !fileExists(fp) {
		logWarning(fp + " not found in project")
		return fp
	}

	switch filepath.Ext(fp) {
	case ".html":
		return handleHTML(fp)
	case ".sass", ".scss":
		return handleSASS(fp)
	default:
		return handleOther(fp)
	}
}

// HTML

func handleHTML(fp string) string {
	finalFilepath := filepath.Join("/",
		strings.TrimSuffix(
			strings.TrimSuffix(fp, "index.html"),
			".html"))

	// Done must be added before handling other *.html files to avoid infinite loop on cyclic references.
	finalPath := path.Clean(path.Join(filepath.SplitList(finalFilepath)...))
	done[fp] = finalPath

	// Step 1: Parse template
	tmpl, data := parseWithLayout(fp, nil)

	// Step 2: Handle other files in generated template
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		logFatal("Cannot execute ", fp, ": ", err)
	}
	doc, err := html.Parse(&b)
	if err != nil {
		panic(err)
	}
	walkNode(doc)

	if err = os.MkdirAll(filepath.Join(outDir, finalFilepath), 0755); err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(outDir, finalFilepath, "index.html"))
	if err != nil {
		panic(err)
	}

	b.Reset()
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
			return template.New("page"), data
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

	page := template.New("page")
	if layout, _ := front["layout"].(string); layout != "" {
		page, data = parseWithLayout(layout, data)
	}
	page = template.Must(page.Parse(string(body)))
	return page, data
}

// walkNode searches for file references in HTML node and replaces them with final filepaths of build.
func walkNode(n *html.Node) {
	if n.Type == html.ElementNode {
		for i, v := range n.Attr {
			if v.Key == "href" || v.Key == "src" {
				n.Attr[i].Val = handleFile(v.Val)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNode(c)
	}
}

// SASS

func handleSASS(fp string) string {
	fb, err := os.ReadFile(fp)
	if err != nil {
		panic(err)
	}

	transpiler, _ := libsass.New(libsass.Options{})
	trans, err := transpiler.Execute(string(fb))
	if err != nil {
		panic(err)
	}

	base := strings.TrimSuffix(fp, filepath.Ext(fp))
	tmpfp := filepath.Join(outDir, base+".css")
	os.WriteFile(tmpfp, []byte(trans.CSS), 0755)
	finalPath := handleOther(tmpfp)
	os.Remove(tmpfp)
	return finalPath
}

// Other file types

func handleOther(fp string) string {
	esLoaderMap := map[string]es.Loader{
		".aac":   es.LoaderFile,
		".avi":   es.LoaderFile,
		".csv":   es.LoaderFile,
		".eot":   es.LoaderFile,
		".gif":   es.LoaderFile,
		".ico":   es.LoaderFile,
		".jpeg":  es.LoaderFile,
		".jpg":   es.LoaderFile,
		".mp3":   es.LoaderFile,
		".mp4":   es.LoaderFile,
		".mpeg":  es.LoaderFile,
		".otf":   es.LoaderFile,
		".png":   es.LoaderFile,
		".pdf":   es.LoaderFile,
		".svg":   es.LoaderFile,
		".ttf":   es.LoaderFile,
		".txt":   es.LoaderFile,
		".webm":  es.LoaderFile,
		".webp":  es.LoaderFile,
		".woff":  es.LoaderFile,
		".woff2": es.LoaderFile,
		".zip":   es.LoaderFile,
	}

	opts := es.BuildOptions{
		EntryPoints:       []string{fp},
		EntryNames:        "[name].[hash]",
		ChunkNames:        "[name].[hash]",
		AssetNames:        "[name].[hash]",
		Outdir:            outDir,
		Bundle:            true,
		Write:             true,
		LogLevel:          es.LogLevelError,
		LegalComments:     es.LegalCommentsNone,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Engines: []es.Engine{
			{Name: api.EngineChrome, Version: "62"},
			{Name: api.EngineFirefox, Version: "69"},
			{Name: api.EngineSafari, Version: "12"},
			{Name: api.EngineEdge, Version: "44"},
		},
		Loader: esLoaderMap,
	}

	res := es.Build(opts)
	if len(res.Errors) > 0 {
		os.Exit(1)
	}

	// Remove unused *.js file generated by esbuild when using the file loader
	if _, ok := esLoaderMap[filepath.Ext(fp)]; ok {
		os.Remove(res.OutputFiles[1].Path)
	}

	// Gerenate final path for HTML use
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	finalFilepath := strings.TrimPrefix(res.OutputFiles[0].Path, filepath.Join(wd, outDir))
	finalPath := path.Clean(path.Join(filepath.SplitList(finalFilepath)...))
	done[fp] = finalPath
	return finalPath
}
