# mdtopdf

[![GoDoc](https://godoc.org/github.com/mandolyte/mdtopdf?status.svg)](https://godoc.org/github.com/mandolyte/mdtopdf)

## Introduction: Markdown to PDF

This package depends on two other packages:
- The BlackFriday v2 parser to read the markdown source
- The `fpdf` packace to generate the PDF

Both of the above are documented at [Go Docs](http://godocs.org).

The tests included here are from the BlackFriday package.
See the "testdata" folder.
The tests create PDF files and thus while the tests may complete
without errors, visual inspection of the created PDF is the
only way to determine if the tests *really* pass!

The tests create log files that trace the BlackFriday parser
callbacks. This is a valuable debug tool showing each callback 
and data provided in each while the AST is presented.

2019-09-23: It appears that the BlackFriday project is no longer active.
There is a fork `gomarkdown/markdown` that I may be able to move to
in the future if needed.

## Supported Markdown
The supported elements of markdown are:
- Emphasized and strong text 
- Headings 1-6
- Ordered and unordered lists
- Nested lists
- Images
- Tables (but see limitations below)
- Links
- Code blocks and backticked text

How to use of non-Latin fonts/languages is documented in a section below.

## Limitations and Known Issues

1. It is common for Markdown to include HTML. HTML is treated as a "code block". *There is no attempt to convert raw HTML to PDF.*

2. Github-flavored Markdown permits strikethough using tildes. This is not supported at present by `fpdf` as a font style.

3. The markdown link title, which would show when converted to HTML as hover-over text, is not supported. The generated PDF will show the actual URL that will be used if clicked, but this is a function of the PDF viewer.

4. Currently all levels of unordered lists use a dash for the bullet. 
This is a planned fix; [see here](https://github.com/mandolyte/mdtopdf/issues/1).

5. Definition lists are not supported (not sure that markdown supports them -- I need to research this)

6. The following text features may be tweaked: font, size, spacing, styile, fill color, and text color. These are exported and available via the `Styler` struct. Note that fill color only works if the text is ouput using CellFormat(). This is the case for: tables, codeblocks, and backticked text.

7. Tables are supported, but no attempt is made to ensure fit. You can, however, change the font size and spacing to make it smaller. See example.



## Installation 

To install the package, run the usual `go get`:
```
go get github.com/mandolyte/mdtopdf
```

## Syntax highlighting

`mdtopdf` supports colourised output via the [gohighlight module](https://github.com/jessp01/gohighlight).

For examples, see `testdata/Markdown Documentation - Syntax.text` and `testdata/Markdown Documentation - Syntax.pdf`

## Quick start

In the `cmd` folder is an example using the package. It demonstrates
a number of features. The test PDF was created with this command:
```
$ go run convert.go -i test.md -o test.pdf
```

To benefit from Syntax highlighting, invoke thusly:

```
$ go run convert.go -i syn_test.md -s /path/to/syntax_files -o test.pdf
```

*Note 1: the `cmd` folder has an example for the syntax highlighting. 
See the script `run_syntax_highlighting.sh`. This example assumes that
the folder with the syntax files is located at relative location:
`../../../jessp01/gohighlight/syntax_files`.*

*Note 2: when annotating the code block to specify the language, the
annotation name must match syntax base filename.*

### Additional options

```sh
  -i string
    	Input text filename; default is os.Stdin
  -o string
    	Output PDF filename; required
  -s string
    	Path to github.com/jessp01/gohighlight/syntax_files
  --new-page-on-hr
    	Interpret HR as a new page; useful for presentations
  --page-size string
    	[A3 | A4 | A5] (default "A4")
  --theme string
    	[light|dark] (default "light")
  --title string
    	Presentation title
  --author string
    	Author; used if -footer is passed
  --font-file string
    	path to font file to use
  --font-name string
    	Font name ID; e.g 'Helvetica-1251'
  --unicode-encoding string
    	e.g 'cp1251'
  --with-footer
    	Print doc footer (author  title  page number)
  --help
    	Show usage message
```

For example, the below will:

- Set the title to `My Grand Title`
- Set `Random Bloke` as the author (used in the footer)
- Set the dark theme
- Start a new page when encountering a HR (`---`); useful for creating presentations
- Print a footer (`author name, title, page number`)

```sh
$ go run convert.go  -i /path/to/md \
    -o apt_dpkg_deb.pdf --title "My Grand Title" --author "Random Bloke" \
    --theme dark --new-page-on-hr --with-footer
```

## Using non-ASCII Glyphs/Fonts

In order to use a non-ASCII language there are a number things that must be done. The PDF generator must be configured WithUnicodeTranslator:

```go
// https://en.wikipedia.org/wiki/Windows-1251
pf := mdtopdf.NewPdfRenderer("", "", *output, "trace.log", mdtopdf.WithUnicodeTranslator("cp1251")) 
```

In addition, this package's `Styler` must be used to set the font to match that is configured with the PDF generator.

A complete working example may be found for Russian in the `cmd` folder nameed
`russian.go`.

For a full example, run:

```sh
$ go run convert.go -i russian.md -o russian.pdf \
    --unicode-encoding cp1251 --font-file helvetica_1251.json --font-name Helvetica_1251
```


# Note to Self

In order to update `pkg.go.dev` with latest release, the following will do the trick. 
Essentially, it is creating a module and then running the go get command for the
desired release.
Using the proxy will have the side effect of updating the info on the go pkg web site.

```sh
$ pwd
/home/cecil/Downloads
$ mkdir tmp
$ cd tmp
$ ls
$ go mod init example.com/mypkg
go: creating new go.mod: module example.com/mypkg
$ cat go.mod 
module example.com/mypkg

go 1.20
$ GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/mandolyte/mdtopdf@v1.4.1
go: added github.com/go-pdf/fpdf v0.8.0
go: added github.com/jessp01/gohighlight v0.21.1-7
go: added github.com/mandolyte/mdtopdf v1.4.1
go: added github.com/russross/blackfriday/v2 v2.1.0
go: added gopkg.in/yaml.v2 v2.4.0
```
