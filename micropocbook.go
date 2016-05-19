package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"

	"github.com/jung-kurt/gofpdf"
)

var logo = flag.String("logo", "chaosvermittlung.png", "Logo to appear at the frontpage")
var nofrontpage = flag.Bool("nofrontpage", false, "Remove frontpage from the PDF")
var event = flag.String("event", "", "Override event string from phonebook from XML")
var font = flag.String("font", "Arial", "Font to use in the phonebook")
var sorting = flag.String("sort", "name", "Defines which data to sort the phonebook. Avaiable options: name, extension")
var title = flag.String("title", "Telefonbuch / Phonebook", "Title of the frontpage")

var width = float64(210)
var height = float64(297)

type entry struct {
	Extension int    `xml:"extension"`
	Name      string `xml:"name"`
}

type ByExentsion []entry

func (a ByExentsion) Len() int           { return len(a) }
func (a ByExentsion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByExentsion) Less(i, j int) bool { return a[i].Extension < a[j].Extension }

type ByName []entry

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type phonebook struct {
	Event   string  `xml:"event"`
	Entries []entry `xml:"entries>entry"`
}

func loadPhonebook(filename string) (phonebook, error) {
	var p phonebook
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return p, err
	}
	err = xml.Unmarshal(body, &p)
	return p, nil
}

func AddNewPage(pdf *gofpdf.Fpdf) {
	pdf.AddPage()
	lm, _, _, _ := pdf.GetMargins()
	pdf.SetXY(lm, 15)
	pdf.SetFont(*font, "", 15)
	pdf.CellFormat(105, 15, "Name", "B", 0, "LM", false, 0, "")
	pdf.CellFormat(0, 15, "Extension", "B", 0, "RM", false, 0, "")
}

func AddEntry(pdf *gofpdf.Fpdf, yoffset float64, xoffset float64, e entry, tr func(string) string) {
	pdf.SetXY(xoffset, yoffset)
	pdf.CellFormat(105, 15, tr(e.Name), "", 0, "LM", false, 0, "")
	pdf.CellFormat(0, 15, strconv.Itoa(e.Extension), "", 0, "RM", false, 0, "")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("No file specified")
	}
	pb, err := loadPhonebook(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	switch *sorting {
	case "name":
		sort.Sort(ByName(pb.Entries))
	case "extension":
		sort.Sort(ByExentsion(pb.Entries))
	default:
		log.Println("Warning: Extension not a vaild option", *sorting)
	}
	if *event != "" {
		pb.Event = *event
	}

	yoffset := float64(30)

	pdf := gofpdf.New("P", "mm", "A4", "")
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	if !*nofrontpage {
		pdf.AddPage()
		imageinfo := pdf.RegisterImage(*logo, "")
		pdf.Image(*logo, 52, yoffset+10, 105, 0, false, "", 0, "")
		texty := (105/imageinfo.Width())*imageinfo.Height() + 30
		pdf.SetFont(*font, "B", 30)
		textx := (width - pdf.GetStringWidth(*title)) / 2
		pdf.Text(textx, yoffset+texty, tr(*title))
		pdf.SetFont(*font, "", 20)
		texty = texty + 20
		textx = (width - pdf.GetStringWidth(pb.Event)) / 2
		pdf.Text(textx, yoffset+texty, tr(pb.Event))
	}
	lm, _, _, bm := pdf.GetMargins()
	AddNewPage(pdf)
	yoffset = 15 + 15
	for i, e := range pb.Entries {
		AddEntry(pdf, yoffset, lm, e, tr)
		yoffset = yoffset + 10
		fmt.Println(i, e)
		if yoffset > (height - bm - 15) {
			AddNewPage(pdf)
			yoffset = 15 + 15
		}
	}

	err = pdf.OutputFileAndClose("phonebook.pdf")
	if err != nil {
		log.Fatal(err)
	}
}
