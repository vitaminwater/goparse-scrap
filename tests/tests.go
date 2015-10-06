package main

import (
	"net/http"
	"fmt"

	"git.ccsas.biz/parse_scrap"
)

func seloger() {
	testpage := "http://www.seloger.com/annonces/achat/appartement/paris-1er-75/101834495.htm?cp=75001&idtt=2&idtypebien=1&listing-listpg=4&tri=d_dt_crea&bd=Li_LienAnn_1"

	scrapper := pscrap.NewScrapper()

	client := &http.Client{}
	page := scrapper.AddPage(client, pscrap.HostMatcher("www.seloger.com"))
	page.AddField(pscrap.FieldPath{"test",}, pscrap.XpathStringArraySelector([]string{"//*[@id=\"slider1\"]/li/img/@src"}))

	if o, err := scrapper.Scrap(testpage, nil); err == nil {
		fmt.Println(o)
	} else {
		panic(err)
	}
}

func lebonCoin() {
	testpage := "http://www.leboncoin.fr/ventes_immobilieres/815473274.htm?ca=12_s"

	scrapper := pscrap.NewScrapper()

	client := &http.Client{}
	page := scrapper.AddPage(client, pscrap.HostMatcher("www.leboncoin.fr"))
	page.AddField(pscrap.FieldPath{"test"}, pscrap.XpathStringArraySelector([]string{"//meta[@itemprop=\"image\"]/@content"}))

	if o, err := scrapper.Scrap(testpage, nil); err == nil {
		fmt.Println(o)
	} else {
		panic(err)
	}
}

func main() {
	//seloger()
	lebonCoin()
}
