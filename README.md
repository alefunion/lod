# Lod

Lod is a static site generator (SSG) with zero configuration.

It only needs an `index.html` entry file.  
Other files are found from there by walking HTML nodes and searching for local file references in `href` or `src` attributes.

- [Install](#install)
- [Usage](#usage)
  - [Build](#build)
  - [Watch and serve](#watch-and-serve)
- [Content types](#content-types)
  - [HTML](#html)
    - [Layout](#layout)
    - [Data](#data)
  - [Styles](#styles)
  - [Scripts](#scripts)
  - [Others](#others)

## Install

```
$ go get github.com/alefunion/lod
```

## Usage

### Build

Go to the project's root directory, write an `index.html` file and run `lod`:

```
$ lod
2021/12/12 12:12:12 ⚡️ SSG in 12ms
```

### Watch and serve

Watching for changes is also possible with the `watch` or `w` subcommand:

```
$ lod w
2021/12/12 12:12:12 ⚡️ SSG in 12ms
2021/12/12 12:12:12 🌐 Server listening on http://localhost:8080
2021/12/12 12:12:12 ⏳ Watching for changes...
```

You can also provide another address to listen on just after the `watch` subcommand:

```
$ lod watch :3000
```

## Content types

### HTML

HTML files are handled as Go templates and can start with a frontmatter.

#### Layout

A layout is an HTML file containing a named `block` (used as a placeholder):

```html
<!DOCTYPE html>
<html>
	<head>
		<title>Example</title>
	</head>
	<body>
		{{ block "main" . }}{{ end }}
	</body>
</html>
```

To use a layout, start the HTML file with a frontmatter containing and a `layout` key referencing the layout file, relative to project's root:

```html
---
layout: layout.html
---

{{define "main"}}
	<h1>Example</h1>
	<p>Lorem ipsum.</p>
{{end}}
```

Layouts can also be nested.

#### Data

Frontmatter's content can be accessed inside the HTML body. It is also passed to layouts:

```html
<!DOCTYPE html>
<html>
	<head>
		<title>Example{{ if .title }} – {{ .title }}{{ end }}</title>
	</head>
	<body>
		<h1>{{ .title }}</h1>
		{{ block "main" . }}{{ end }}
	</body>
</html>
```

```html
---
layout: layout.html
title: Test
---

{{define "main"}}
	<p>{{ .title }} page</p>
{{end}}
```

### Styles

CSS files are automatically minified and bundled.  
The official Sass/Scss compiler is alos built in.  
Just make a reference to the entry file:

```html
<link rel="stylesheet" href="main.scss" />
```

### Scripts

JS and TS files are automatically minified and bundled.  
As they are handled by [esbuild](https://esbuild.github.io), you can use the latest syntax, like ES6 imports.

```html
<script src="scripts/main.js"></script>
```

### Others

Other content types are copied as is with a hashed filename.
