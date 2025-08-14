package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type jobsScrap struct {
	Title    string
	Salary   string
	Tags     []string
	Company  string
	Location string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var htmlContent string

	urlScrap := "https://glints.com/id/opportunities/jobs/explore?keyword=backend&country=ID&locationName=All+Cities%2FProvinces&lowestLocationLevel=1"
	log.Println("Menjalankan Chromedp")

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(urlScrap),
		chromedp.WaitVisible(`.JobCardsc__JobcardContainer-sc-hmqj50-0`, chromedp.ByQuery),
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatalf("gagal run chromedp %v \n", err)
	}

	log.Println("Sukses")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))

	rows := make([]jobsScrap, 0)

	doc.Find(`.CompactOpportunityCardsc__CompactJobCard-sc-dkg8my-4`).Each(func(i int, s *goquery.Selection) {
		row := new(jobsScrap)
		tags := make([]string, 0)

		var cleanItem = func(s string) string {

			var b strings.Builder
			b.Grow(len(s)) // Optimasi: alokasikan memori seukuran string awal

			for _, r := range s {
				// Hanya simpan karakter yang merupakan Huruf, Angka, Tanda Baca, Simbol, atau Spasi biasa.
				if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || r == ' ' {
					b.WriteRune(r)
				}
			}

			// Setelah difilter, lakukan normalisasi spasi sekali lagi untuk hasil akhir yang sempurna.
			cleaned := b.String()

			// LANGKAH 1: Normalisasi karakter-karakter tipografi spesifik
			// Ganti en dash (–) dengan hyphen biasa (-)
			cleaned = strings.ReplaceAll(cleaned, "–", "-")
			// Bonus: Ganti juga em dash (—), yang lebih panjang
			cleaned = strings.ReplaceAll(cleaned, "—", "-")
			// Bonus: Ganti juga karakter elipsis (...) unicode dengan tiga titik biasa
			cleaned = strings.ReplaceAll(cleaned, "…", "...")

			fields := strings.Fields(cleaned)
			return strings.Join(fields, " ")
		}

		rawTitle := s.Find(".CompactOpportunityCardsc__JobCardTitleNoStyleAnchor-sc-dkg8my-12").Text()
		row.Title = cleanItem(rawTitle)

		rawSalary := s.Find(".CompactOpportunityCardsc__SalaryWrapper-sc-dkg8my-32").Text()
		row.Salary = cleanItem(rawSalary)
		// row.tags = s.Find().Text()
		s.Find(".TagStyle-sc-r1wv7a-4").Each(func(i int, sel *goquery.Selection) {
			// getTag := new(string)
			getTag := sel.Find(".TagStyle__TagContentWrapper-sc-r1wv7a-1").Text()
			cleanedTags := cleanItem(getTag)
			tags = append(tags, cleanedTags)
		})
		row.Tags = tags

		rawCompany := s.Find(".CompactOpportunityCardsc__CompanyLink-sc-dkg8my-14").Text()
		row.Company = cleanItem(rawCompany)

		rawLocation := s.Find(".CardJobLocation__LocationWrapper-sc-v7ofa9-0").Text()
		row.Location = cleanItem(rawLocation)

		rows = append(rows, *row)
	})

	bts, err := json.MarshalIndent(rows, "", " ")
	if err != nil {
		log.Fatalln("error ketika marshal indent ", err)
	}

	// log.Println(string(bts))

	filename := "result.json"
	os.WriteFile(filename, bts, 0664)
}
