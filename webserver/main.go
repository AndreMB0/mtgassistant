// program webserver is a web server that parses out MTG Arena Logs and outputs the contents of the cards boosters that you opened
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

        "github.com/andremb0/mtgassistant/carddb"
	"github.com/andremb0/mtgassistant/collectionfinder"
)

var (
	mtgDataPath = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	landingpage = flag.String("landing", "trackboosters", "Path to the landing page.")
	mtgSet      = flag.String("set", "M21", "Current Set for wildcard picking.")
	jsonFormat  = flag.Bool("json", false, "Whether or not to output booster info in JSON format.")
)

var templates = template.Must(template.ParseFiles("trackboosters.html", "view.html"))
var boosterData []collectionfinder.BoosterContents
var cardsByRarity = make(map[int][]uint64)

const maxMtgaLogsSize int64 = 100 << 20 // 20 MiB

// BoosterContents represent the contents of a MTG:Arena booster pack.
type BoosterContents struct {
	WcCommon   int      `json:"wcc"`
	WcUncommon int      `json:"wcu"`
	WcRare     int      `json:"wcr"`
	WcMythic   int      `json:"wcm"`
	Cards      []string `json:"cards"`
}

// Page struct
type Page struct {
	Title          string
	Body           []byte
	BoosterTextMap map[int][]string
	totalWCMap     map[uint64]int
	CardsMap       map[uint64]int
	SummaryText    string
	WildcardText   string
	OnlyWCText     string
}

func loadPage(title string) (*Page, error) {
	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to parse %s page file: %v", title, err)
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func outputJSON(w http.ResponseWriter, boosterData []collectionfinder.BoosterContents) {
	var boosters []BoosterContents

	w.Header().Add("Content-Type", "application/json")

	for _, booster := range boosterData {
		var contents BoosterContents
		contents.WcCommon = booster.CommonWildcards
		contents.WcUncommon = booster.UncommonWildcards
		contents.WcRare = booster.RareWildcards
		contents.WcMythic = booster.MythicWildcards
		for _, id := range booster.CardIds {
			card := db.GetCardByID(id)
			c := fmt.Sprintf("%d %s (%s) %s", 1, card.Name, card.Set, card.CollectorNumber)
			contents.Cards = append(contents.Cards, c)
		}
		boosters = append(boosters, contents)
	}

	enc := json.NewEncoder(w)
	enc.Encode(boosters)
}

func calculateWCPicker(p *Page) string {
	var onlyWCText = ""
	var hashNum = int64(0)
	var randomCard uint64
	for _, id := range boosterData[0].CardIds {
		hashNum += int64(id)
	}
	rand.Seed(hashNum)
	for i := 1; i <= p.totalWCMap[carddb.CommonRarity]; i++ {
		randomCard = uint64(rand.Int() % len(cardsByRarity[carddb.CommonRarity]))
		wcCardID := cardsByRarity[carddb.CommonRarity][randomCard]
		p.CardsMap[wcCardID]++
		card := db.GetCardByID(wcCardID)
		log.Printf("calculating common wc: %d with random number %d and hashnum %d", wcCardID, randomCard, hashNum)
		log.Printf("card %s", card.Name)
		onlyWCText += fmt.Sprintf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
	}
	for i := 1; i <= p.totalWCMap[carddb.UncommonRarity]; i++ {
		randomCard = uint64(rand.Int() % len(cardsByRarity[carddb.UncommonRarity]))
		wcCardID := cardsByRarity[carddb.UncommonRarity][randomCard]
		p.CardsMap[wcCardID]++
		card := db.GetCardByID(wcCardID)
		log.Printf("calculating uncommon wc: %d with random number %d and hashnum %d", wcCardID, randomCard, hashNum)
		log.Printf("card %s", card.Name)
		onlyWCText += fmt.Sprintf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
	}
	for i := 1; i <= p.totalWCMap[carddb.RareRarity]; i++ {
		randomCard = uint64(rand.Int() % len(cardsByRarity[carddb.RareRarity]))
		wcCardID := cardsByRarity[carddb.RareRarity][randomCard]
		p.CardsMap[wcCardID]++
		card := db.GetCardByID(wcCardID)
		log.Printf("calculating rare wc: %d with random number %d and hashnum %d", wcCardID, randomCard, hashNum)
		log.Printf("card %s", card.Name)
		onlyWCText += fmt.Sprintf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
	}
	for i := 1; i <= p.totalWCMap[carddb.MythicRarity]; i++ {
		randomCard = uint64(rand.Int() % len(cardsByRarity[carddb.MythicRarity]))
		wcCardID := cardsByRarity[carddb.MythicRarity][randomCard]
		p.CardsMap[wcCardID]++
		card := db.GetCardByID(wcCardID)
		log.Printf("calculating mythic wc: %d with random number %d and hashnum %d", wcCardID, randomCard, hashNum)
		log.Printf("card %s", card.Name)
		onlyWCText += fmt.Sprintf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
	}

	return onlyWCText
}

func calculateCardsByRarity(c carddb.Card) {
	for rarity := 2; rarity <= 5; rarity++ {
		if c.Rarity == uint64(rarity) && c.Set == *mtgSet && c.IsPrimaryCard {
			// log.Printf("card: %s", c.Name)
			cardsByRarity[rarity] = append(cardsByRarity[rarity], c.ID)
		}
	}
}

func getPageContent() *Page {
	var p Page
	var cardsMap = make(map[uint64]int)

	var totalWCMap = make(map[uint64]int)
	var BoosterMap = make(map[int][]string)
	for k := range cardsByRarity {
		delete(cardsByRarity, k)
	}

	for i, booster := range boosterData {
		var BText = ""

		for _, id := range booster.CardIds {
			card := db.GetCardByID(id)
			BText += fmt.Sprintf("%d %s (%s) %s\n", 1, card.Name, card.Set, card.CollectorNumber)
			cardsMap[id]++
		}

		BText += fmt.Sprintf("\n\n\nCommon Wildcards: %d\nUncommon Wildcards: %d\nRare Wildcards: %d\nMythic Wildcards: %d\n",
			booster.CommonWildcards, booster.UncommonWildcards, booster.RareWildcards, booster.MythicWildcards)
		BoosterMap[i] = append(BoosterMap[i], BText)

		totalWCMap[carddb.CommonRarity] += booster.CommonWildcards
		totalWCMap[carddb.UncommonRarity] += booster.UncommonWildcards
		totalWCMap[carddb.RareRarity] += booster.RareWildcards
		totalWCMap[carddb.MythicRarity] += booster.MythicWildcards
	}
	p.WildcardText += fmt.Sprintf("\nTOTAL:\nCommon Wildcards: %d\nUncommon Wildcards: %d\nRare Wildcards: %d\nMythic Wildcards: %d\n",
		totalWCMap[carddb.CommonRarity], totalWCMap[carddb.UncommonRarity], totalWCMap[carddb.RareRarity], totalWCMap[carddb.MythicRarity])

	p.BoosterTextMap = BoosterMap
	p.totalWCMap = totalWCMap
	p.CardsMap = cardsMap
	return &p
}

func outputPlain(w http.ResponseWriter, boosterData []collectionfinder.BoosterContents, p *Page) {
	w.Header().Add("Content-Type", "text/html")
	var Summary = ""

	p = getPageContent()

	for key, value := range p.CardsMap {
		card := db.GetCardByID(key)
		Summary += fmt.Sprintf("%d %s (%s) %s\n", value, card.Name, card.Set, card.CollectorNumber)
	}

	p.SummaryText = Summary

	renderTemplate(w, "view", p)
}

func wcPickerHandler(dc carddb.CardDB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var Summary = ""

		p, err := loadPage("view")
		if err != nil {
			http.Redirect(w, r, "/view", http.StatusFound)
			return
		}
		w.Header().Add("Content-Type", "application/json")

		p = getPageContent()

		db.ForEach(calculateCardsByRarity)

		p.OnlyWCText = calculateWCPicker(p)

		for key, value := range p.CardsMap {
			card := db.GetCardByID(key)
			Summary += fmt.Sprintf("%d %s (%s) %s\n", value, card.Name, card.Set, card.CollectorNumber)
		}

		p.SummaryText = Summary

		pageJSON, err := json.Marshal(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(pageJSON)
	}
}

func uploadHandler(dc carddb.CardDB, jsonFormat bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxMtgaLogsSize); err != nil {
			log.Printf("could not parse multipart form: %v", err)
			http.Error(w, "Invalid Request", http.StatusPreconditionFailed)
		}

		file, _, err := r.FormFile("mtgalogs")
		if err != nil {
			log.Printf("couldnt get uploaded file: %v", err)
			http.Error(w, "Could not retrieve mtg logs file", http.StatusPreconditionFailed)
			return
		}
		defer file.Close()
		boosterData, err = collectionfinder.FindBoosters(file)
		if err != nil {
			log.Fatalf("failed to parse mtga logs: %v", err)
		}

		p, err := loadPage("view")
		if err != nil {
			http.Redirect(w, r, "/view", http.StatusFound)
			return
		}

		if jsonFormat {
			outputJSON(w, boosterData)
		} else {
			outputPlain(w, boosterData, p)
		}
	}
}

func boosterTracker(landingpagedata *Page) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingpagedata.Body)
	}
}

var db carddb.CardDB

func main() {
	flag.Parse()

	log.Println("Parsing MTG Data Files...")

	landingpagedata, err := loadPage(*landingpage)

	db, err = carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	log.Println("Starting Server")
	http.HandleFunc("/wildcardpicker", wcPickerHandler(db))
	http.HandleFunc("/view", uploadHandler(db, *jsonFormat))
	http.HandleFunc("/trackboosters", boosterTracker(landingpagedata))
	http.Handle("/", http.FileServer(http.Dir(".")))
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}

