package main

import (
	"net/http"
	"golang.org/x/net/html"
	"github.com/yhat/scrape"
	"fmt"
	"strings"
	"regexp"
	"strconv"
)

type SiteLocation interface {
	getAnnonce(n *html.Node) bool
	getNext(n *html.Node) bool
	getSurface(n *html.Node) bool
	getPrix(n *html.Node) bool
	getPieces(n *html.Node) bool
	getType(n *html.Node) bool
	crawl(n *html.Node)
}

type requete struct {
	url, prixMin, prixMax,pieceMin, pieceMax, surfacemin, surfaceMax, isMeuble string
}

type requeteMaker interface {
	setPrix(int, int) // Lower and higher
	setPiece(int, int) // Lower and higher
	setSurface(int, int) // Lower adn higher
	isMeuble(int)
	make()
}




type location struct {
	titre, typeLoca string
	prix, surface, pieces int
}

func (l location) getAnnonce(n *html.Node) bool{return false}
func (l location) getNext(n *html.Node) bool{return false}
func (l location) getSurface(n *html.Node) bool{return false}
func (l location) getPrix(n *html.Node) bool{return false}
func (l location) getPieces(n *html.Node) bool{return false}
func (l location) getType(n *html.Node) bool{return false}
func (l location) crawl(n *html.Node){}

func (l location) String() (retr string) {
	retr = l.titre + "\n" + l.typeLoca + ", " + strconv.Itoa(l.pieces) + ", " + strconv.Itoa(l.surface) + "m2\n" + strconv.Itoa(l.prix) + "€\n"
	return
}

type leBonCoin struct {
	location
}

func (lbc leBonCoin) getAnnonce (n *html.Node) bool {
	class := scrape.Attr(n, "class")
	return strings.Contains(class, "list_item")
}

func(lbc leBonCoin) getNext (n *html.Node) bool {
	return scrape.Attr(n, "id") == "next"
}

func (lbc leBonCoin) getSurface (n *html.Node) bool {
	return scrape.Attr(n, "data-qa-id") == "criteria_item_square"
}

func (lbc leBonCoin) getPrix (n *html.Node) bool {
	return scrape.Attr(n, "data-qa-id") == "adview_price"
}

func (lbc leBonCoin) getPieces (n *html.Node) bool {
	return scrape.Attr(n, "data-qa-id") == "criteria_item_rooms"
}

func (lbc leBonCoin) getType (n *html.Node) bool {
	return scrape.Attr(n, "data-qa-id") == "criteria_item_real_estate_type"
}

func (lbc *leBonCoin) crawl(annonce *html.Node) {

	var validID = regexp.MustCompile(`^\d*`)

	urlAnnonce := "https:" + scrape.Attr(annonce, "href")
	fmt.Println(urlAnnonce)
	pageAnnonce, _ := http.Get(urlAnnonce)
	rootAnnonce, _ := html.Parse(pageAnnonce.Body)

	surfaceRect, ok := scrape.Find(rootAnnonce, lbc.getSurface)
	if ok {
		surface := surfaceRect.LastChild.LastChild.FirstChild.Data
		surface = validID.FindString(surface)
		lbc.surface, _ = strconv.Atoi(surface)
	} else {
		lbc.surface = -12
	}

	piecesRect, ok := scrape.Find(rootAnnonce, lbc.getPieces)
	if ok {
		pieces := piecesRect.LastChild.LastChild.FirstChild.Data
		lbc.pieces, _ = strconv.Atoi(pieces)
	} else {
		lbc.pieces = -12
	}

	prixRect, ok := scrape.Find(rootAnnonce, lbc.getPrix)
	if ok {
		prix := scrape.Text(prixRect)
		prix = validID.FindString(prix)
		lbc.prix, _ = strconv.Atoi(prix)
	} else {
		lbc.prix = -12
	}

	typeRect, ok := scrape.Find(rootAnnonce, lbc.getType)
	if ok {
		lbc.typeLoca = typeRect.LastChild.LastChild.FirstChild.Data
	} else {
		lbc.typeLoca = ""
	}

	lbc.titre = scrape.Attr(annonce, "title")
}

func main() {


	/** TODO : Lecture des parametre de la requete a envoyée
	** Vile, Prix minimum, prix maximum, meublé ou pas, ... 
	**/

	/*
	var req string
	fmt.Println("Saisir le nom de la ville")
	fmt.Scanln(req)
	fmt.Println(req)
	*/


	var url string = "https://www.leboncoin.fr/locations/offres?th=1&location=Evry&mre=550&sqs=5"

	for {
		resp, err := http.Get(url)
		if err != nil {
			return
		}
		root, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}

		var lbc leBonCoin

		// Grab all annonces and print them
		annonces := scrape.FindAll(root, lbc.getAnnonce)
		for _, annonce := range annonces {
			lbc.crawl(annonce)
			fmt.Println(lbc)
		}

		nextPage, ok := scrape.Find(root, lbc.getNext)

		if ok {
			url = "https:" + scrape.Attr(nextPage, "href")
		} else {
			fmt.Println("Fin de la recherche")
			return
		}
	}
}
