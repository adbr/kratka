// 2016-12-02 adbr

// Program kratka tworzy plik pdf zawierający jedną stronę w kratkę.
// Opis programu znajduje się w 'const helpStr' lub może być
// wyświetlony przy użyciu opcji -h.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// Prefiks nazwy plików roboczych
const basename = "kratka"

// Parametry dokumentu w latexu używane w template
type Parameters struct {
	Margin    string
	Hoffset   string
	Voffset   string
	Showframe bool
	BoxSizeX  string
	BoxSizeY  string
	Step      string
	LineWidth string
	LineColor string
	LineStyle string
	GridSizeX int
	GridSizeY int
}

// Flagi command line
var (
	work      = flag.Bool("work", false, "")
	help      = flag.Bool("h", false, "")
	tmpl      = flag.Bool("template", false, "")
	margin    = flag.String("margin", "1cm", "")
	hoffset   = flag.String("hoffset", "0cm", "")
	voffset   = flag.String("voffset", "0cm", "")
	showframe = flag.Bool("showframe", false, "")
	boxsizex  = flag.String("boxsizex", "4.25mm", "")
	boxsizey  = flag.String("boxsizey", "4.25mm", "")
	step      = flag.String("step", "4.25mm", "")
	linewidth = flag.String("linewidth", "very thin", "")
	linecolor = flag.String("linecolor", "gray", "")
	linestyle = flag.String("linestyle", "solid", "")
	gridsizex = flag.Int("gridsizex", 43, "")
	gridsizey = flag.Int("gridsizey", 64, "")
)

func main() {
	log.SetFlags(0) // nie drukuj daty i czasu
	log.SetPrefix("kratka: ")

	flag.Usage = usage
	flag.Parse()

	if *help {
		fmt.Print(helpStr)
		os.Exit(0)
	}

	if *tmpl {
		fmt.Print(latexTemplate)
		os.Exit(0)
	}

	// Sprawdzenie czy podano nazwę pliku wynikowego
	if flag.NArg() < 1 {
		log.Print("brakuje argumentu z nazwą pliku pdf")
		usage()
		os.Exit(2)
	}

	// Utworzenie katalogu roboczego
	workdir, err := ioutil.TempDir("", basename)
	if err != nil {
		log.Fatal(err)
	}

	// Utworzenie pliku w latexu
	fname := createLatexFile(workdir)

	// Kompilacja pliku w latexu i wygenerowanie pliku pdf
	cmd := exec.Command("pdflatex", "-output-directory", workdir,
		"-halt-on-error", fname)
	err = cmd.Run()
	if err != nil {
		logfile := filepath.Join(workdir, basename+".log")
		log.Print("pdflatex error")
		log.Printf("pdflatex log file: %s", logfile)
		log.Fatal(err)
	}

	// Skopiowanie wynikowego pliku pdf
	outfile := flag.Arg(0)
	pdffile := filepath.Join(workdir, basename+".pdf")
	err = copyFile(outfile, pdffile)
	if err != nil {
		log.Fatalf("copy pdf file: %s", err)
	}

	// Usunięcie katalogu roboczego
	if *work {
		log.Printf("work directory: %s\n", workdir)
	} else {
		err = os.RemoveAll(workdir)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// createLatexFile generuje w katalogu workdir plik w latexu używając
// template. Zwraca nazwę utworzonego pliku. W przypadku błedu kończy
// działanie programu.
func createLatexFile(workdir string) (fname string) {
	fname = filepath.Join(workdir, basename+".tex")
	file, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	tmpl := template.New("latex").Delims("@@", "@@")
	tmpl, err = tmpl.Parse(latexTemplate)
	if err != nil {
		log.Fatal(err)
	}

	// Ustawienie parametrów dokumentu na podstawie flag
	var params Parameters
	params.Margin = *margin
	params.Hoffset = *hoffset
	params.Voffset = *voffset
	params.Showframe = *showframe
	params.BoxSizeX = *boxsizex
	params.BoxSizeY = *boxsizey
	params.Step = *step
	params.LineWidth = *linewidth
	params.LineColor = *linecolor
	params.LineStyle = *linestyle
	params.GridSizeX = *gridsizex
	params.GridSizeY = *gridsizey

	// Wygenerowanie pliku
	err = tmpl.Execute(file, params)
	if err != nil {
		log.Fatal(err)
	}

	return fname
}

// copyFile kopiuje plik src do pliku dst. Plik docelowy dst jest
// nadpisywany jeśli istnieje. Jeśli nazwą pliku dst jest "-" to
// kopiuje do os.Stdout.
func copyFile(dst, src string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	var out *os.File
	if dst == "-" {
		out = os.Stdout
	} else {
		out, err = os.Create(dst)
		if err != nil {
			return err
		}
		defer func() {
			e := out.Close()
			if err == nil && e != nil {
				err = e
			}
		}()
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	if dst != "-" {
		err = out.Sync()
		if err != nil {
			return err
		}
	}

	return nil
}

func usage() {
	fmt.Fprint(os.Stderr, usageStr)
}

// latexTemplate zawiera template dokumentu w latexu. Ogranicznikami
// dla template actions są @@ i @@ (zamiast domyślnych {{ i }}).
const latexTemplate = `
\documentclass[a4paper,11pt]{memoir}

\usepackage[MeX]{polski}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage[margin=@@.Margin@@,
  hoffset=@@.Hoffset@@,
  voffset=@@.Voffset@@,
  showframe=@@.Showframe@@]{geometry}
\usepackage{tikz}
\usepackage{layout}

\setlength{\topskip}{0mm}
\setlength{\parindent}{0mm}

\begin{document}

%\layout
\pagestyle{empty}

\begin{vplace}
  \begin{centering}

    \begin{tikzpicture}[x=@@.BoxSizeX@@, y=@@.BoxSizeY@@]
      \draw[step=@@.Step@@, @@.LineWidth@@, @@.LineColor@@, @@.LineStyle@@]
        (0,0) grid (@@.GridSizeX@@, @@.GridSizeY@@);
    \end{tikzpicture}

  \end{centering}
\end{vplace}

\end{document}
`

const usageStr = `Sposób użycia: kratka [opcje] plik.pdf
Opcje:
	-work
		drukuje nazwę tymczasowego katalogu roboczego i nie
		usuwa go na końcu.
	-h
		drukuje help
	-template
		drukuje template dokumentu w latexu
	-margin length
		margines dookoła rysunku (domyślnie "1cm")
	-hoffset length
		przesunięcie rysunku w poziomie (domyślnie "0cm")
	-voffset length
		przesunięcie rysunku w pionie (domyślnie "0cm")
	-showframe
		czy rysować linie odniesienia (domyślnie false)
	-boxsizex length
		rozmiar kratki w poziomie (domyślnie "4.25mm")
	-boxsizey length
		rozmiar kratki w pionie (domyślnie "4.25mm")
	-step length
		(domyślnie "4.25mm")
	-linewidth width
		grubość linii (domyślnie "very thin")
	-linecolor color
		kolor linii (domyślnie "gray")
	-linestyle style
		rodzaj linii (domyślnie "solid")
	-gridsizex int
		liczba kratek w poziomie (domyślnie 43)
	-gridsizey int
		liczba kratek w pionie (domyślnie 64)
`

const helpStr = `Program kratka tworzy plik pdf zawierający jedną stronę w kratkę.

Plik pdf jest tworzony przy użyciu programu pdflatex z systemu LaTeX.
Domyślnie jest tworzona strona z kratką 43x64 pola o rozmiarze 4.25mm.

Sposób użycia: kratka [opcje] plik.pdf
Opcje:
	-work
		drukuje nazwę tymczasowego katalogu roboczego i nie
		usuwa go na końcu.
	-h
		drukuje help
	-template
		drukuje template dokumentu w latexu
	-margin length
		margines dookoła rysunku (domyślnie "1cm")
	-hoffset length
		przesunięcie rysunku w poziomie (domyślnie "0cm")
	-voffset length
		przesunięcie rysunku w pionie (domyślnie "0cm")
	-showframe
		czy rysować linie odniesienia (domyślnie false)
	-boxsizex length
		rozmiar kratki w poziomie (domyślnie "4.25mm")
	-boxsizey length
		rozmiar kratki w pionie (domyślnie "4.25mm")
	-step length
		(domyślnie "4.25mm")
	-linewidth width
		grubość linii (domyślnie "very thin")
	-linecolor color
		kolor linii (domyślnie "gray")
	-linestyle style
		rodzaj linii (domyślnie "solid")
	-gridsizex int
		liczba kratek w poziomie (domyślnie 43)
	-gridsizey int
		liczba kratek w pionie (domyślnie 64)

Argument length w opcjach ma format taki jak length w LaTeXu.
Przykłady wartości argumentu length:

	1cm, -2.34cm, 3.0mm, 4pt, 5in, 6ex, 7em

Argument width opcji -linewidth może mieć wartość taką jak grubości
linii w latexowym pakiecie TikZ. Przykłady wartości argumentu width:

	line width=5pt
	ultra thin
	very thin
	thin
	semithick
	thick
	very thick
	ultra thick

Argument color opcji -linecolor może mieć wartość taką jak kolor
linii w latexowym pakiecie TikZ. Przykłady wartości argumentu color:

	gray, blue, red

Argument style opcji -linestyle może mieć wartość taką jak rodzaj
linii w latexowym pakiecie TikZ. Przykłady wartości argumentu style:

	solid
	dotted
	densely dotted
	dashed
	densely dashed
	dash dot
	dash dot dot

Przykłady:

Kratka o domyślnych parametrach ale rysowana liniami kropkowanymi:

	kratka -linestyle dotted x.pdf

Kratka o domyślnych parametrach przesunięta o 2 mm w lewo:

	kratka -hoffset -2mm x.pdf

Zależności:

Do działania programu musi być zainstalowany system LaTeX z pakietami:
- memoir
- geometry
- tikz
W ścieżce musi być dostępny program pdflatex.
`
