package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gomarkdown/markdown/parser"
	"github.com/mandolyte/mdtopdf"
	"golang.org/x/exp/slices"
)

type FontFileList struct {
	files []string
}

func (l *FontFileList) String() string {
	return strings.Join(l.files, ",")
}

func (l *FontFileList) Set(v string) error {
	q := strings.SplitAfter(v, ",")
	l.files = append(l.files, q...)
	return nil
}

func (l *FontFileList) Count() int {
	return len(l.files)
}

var fontfiles = FontFileList{
	files: []string{},
}

var input = flag.String("i", "", "Input filename, dir consisting of .md|.markdown files or HTTP(s) URL; default is os.Stdin")
var output = flag.String("o", "", "Output PDF filename; required")
var pathToSyntaxFiles = flag.String("s", "", "Path to github.com/jessp01/gohighlight/syntax_files")
var title = flag.String("title", "", "Presentation title")
var author = flag.String("author", "", "Author; used if -footer is passed")
var unicodeSupport = flag.String("unicode-encoding", "", "e.g 'cp1251'")
var fontName = flag.String("font-name", "", "Font name ID; e.g 'Helvetica-1251'")
var themeArg = flag.String("theme", "light", "[light|dark]")
var hrAsNewPage = flag.Bool("new-page-on-hr", false, "Interpret HR as a new page; useful for presentations")
var printFooter = flag.Bool("with-footer", false, "Print doc footer (author  title  page number)")
var pageSize = flag.String("page-size", "A4", "[A3 | A4 | A5]")
var orientation = flag.String("orientation", "portrait", "[portrait | landscape]")
var logFile = flag.String("log-file", "", "Path to log file")
var help = flag.Bool("help", false, "Show usage message")
var version = "dev"
var _, fileName, fileLine, ok = runtime.Caller(0)

var opts []mdtopdf.RenderOption

func processRemoteInputFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code: " + fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	content, rerr := io.ReadAll(resp.Body)
	return content, rerr
}

func glob(dir string, validExts []string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if slices.Contains(validExts, filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Var(&fontfiles, "font-file", "path to font file(s) to use, comma-separated")
	flag.Parse()

	if *help {
		usage("Help Message")
	}

	if *output == "" {
		usage("Output PDF filename is required")
	}

	if *hrAsNewPage == true {
		opts = append(opts, mdtopdf.IsHorizontalRuleNewPage(true))
	}

	if *unicodeSupport != "" {
		opts = append(opts, mdtopdf.WithUnicodeTranslator(*unicodeSupport))
	}

	if *pathToSyntaxFiles != "" {
		opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir(*pathToSyntaxFiles))
	} else {
		opts = append(opts, mdtopdf.SetSyntaxHighlightBaseDir("../../highlight/syntax_files"))
	}

	// get text for PDF
	var content []byte
	var err error
	var inputBaseURL string
	if *input == "" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		httpRegex := regexp.MustCompile("^http(s)?://")
		if httpRegex.Match([]byte(*input)) {
			content, err = processRemoteInputFile(*input)
			if err != nil {
				log.Fatal(err)
			}
			// get the base URL so we can adjust relative links and images
			inputBaseURL = strings.Replace(filepath.Dir(*input), ":/", "://", 1)
		} else {
			fileInfo, err := os.Stat(*input)
			if err != nil {
				log.Fatal(err)
			}

			if fileInfo.IsDir() {
				opts = append(opts, mdtopdf.IsHorizontalRuleNewPage(true))
				validExts := []string{".md", ".markdown"}
				files, err := glob(*input, validExts)
				if err != nil {
					log.Fatal(err)
				}
				for i, filePath := range files {
					fileContents, err := os.ReadFile(filePath)
					if err != nil {
						log.Fatal(err)
					}
					content = append(content, fileContents...)
					if i < len(files)-1 {
						content = append(content, []byte("---\n")...)
					}
				}
			} else {
				content, err = os.ReadFile(*input)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	theme := mdtopdf.LIGHT
	textColor := mdtopdf.Colorlookup("black")
	fillColor := mdtopdf.Colorlookup("white")
	backgroundColor := "white"
	if *themeArg == "dark" {
		theme = mdtopdf.DARK
		backgroundColor = "black"
		textColor = mdtopdf.Colorlookup("darkgray")
		fillColor = mdtopdf.Colorlookup("black")
	}

	pf := mdtopdf.NewPdfRenderer(*orientation, *pageSize, *output, *logFile, opts, theme)
	if inputBaseURL != "" {
		pf.InputBaseURL = inputBaseURL
	}
	pf.Pdf.SetSubject(*title, true)
	pf.Pdf.SetTitle(*title, true)
	pf.BackgroundColor = mdtopdf.Colorlookup(backgroundColor)
	pf.Extensions = parser.NoIntraEmphasis | parser.Tables | parser.FencedCode | parser.Autolink | parser.Strikethrough | parser.SpaceHeadings | parser.HeadingIDs | parser.BackslashLineBreak | parser.DefinitionLists

	useCustomFont := false

	if fontfiles.Count() > 0 && *fontName != "" {
		for _, n := range fontfiles.files {
			flag := ""
			if true == strings.Contains(n, "Bold") {
				flag = "B"
			} else if true == strings.Contains(n, "Italic") {
				flag = "I"
			}

			pf.Pdf.AddFont(*fontName, flag, n)
		}

		pf.Pdf.SetFont(*fontName, "", 12)
		useCustomFont = true
		pf.Normal = mdtopdf.Styler{Font: *fontName, Style: "",
			Size: 12, Spacing: 2,
			FillColor: fillColor,
			TextColor: textColor}
		pf.SetCustomFont(*fontName)
		pf.UpdateParagraphStyler(pf.Normal)
	}

	if *printFooter {
		pf.Pdf.SetFooterFunc(func() {
			color := mdtopdf.Colorlookup(backgroundColor)
			pf.Pdf.SetFillColor(color.Red, color.Green, color.Blue)
			// Position at 1.5 cm from bottom
			pf.Pdf.SetY(-15)
			if useCustomFont == true {
				pf.Pdf.SetFont(*fontName, "I", 8)
			} else {
				// Arial italic 8
				pf.Pdf.SetFont("Arial", "I", 8)
			}
			// Text color in gray
			pf.Pdf.SetTextColor(128, 128, 128)
			w, h, _ := pf.Pdf.PageSize(pf.Pdf.PageNo())
			// fmt.Printf("Width: %f, height: %f, unit: %s\n", w, h, u)
			pf.Pdf.SetX(4)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *author), "", 0, "", true, 0, "")
			middle := w / 2
			if *orientation == "landscape" {
				middle = h / 2
			}
			pf.Pdf.SetX(middle - float64(len(*title)))
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("%s", *title), "", 0, "", true, 0, "")
			pf.Pdf.SetX(-40)
			pf.Pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pf.Pdf.PageNo()), "", 0, "", true, 0, "")
		})
	}

	err = pf.Process(content)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func usage(msg string) {
	fmt.Println(msg + "\n")
	fmt.Printf("Usage: %s (%s) [options]\n", filepath.Base(fileName), version)
	flag.PrintDefaults()
	os.Exit(0)
}
