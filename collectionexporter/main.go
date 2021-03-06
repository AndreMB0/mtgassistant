// program collectionexporter parses a "Magic The Gathering - Arena" output log and prints out the card collectioon of the user.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mvanotti/mtgassistant/carddb"
	"github.com/mvanotti/mtgassistant/collectionfinder"
)

var (
	mtgOutputLog = flag.String("log_file", `${USERPROFILE}\AppData\LocalLow\Wizards Of The Coast\MTGA\output_log.txt`, "Filepath of the MTG Arena Output Log, typically stored in an MTG folder inside C:\\Users")
	mtgDataPath  = flag.String("mtg_data", `C:\Program Files (x86)\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data`, "Path to the Downloads\\Data folder inside the MTG Arena Install Directory")
	inventory    = flag.Bool("inventory", false, "Also output user inventory")
)

func main() {
	flag.Parse()
	log.Println("Parsing MTGA Log...")
	f, err := os.Open(os.ExpandEnv(*mtgOutputLog))
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()
	cardLists, err := collectionfinder.FindCollection(f)
	if err != nil {
		log.Fatalf("failed to parse mtga logs: %v", err)
	}
	if len(cardLists) < 1 {
		log.Fatal("no decks found in the mtg logs. make sure to enable logs in the Arena app.")
	}
	cardList := cardLists[len(cardLists)-1]

	log.Println("Parsing MTG Data Files...")
	db, err := carddb.CreateLibrary(*mtgDataPath)
	if err != nil {
		log.Fatalf("createLibrary failed: %v", err)
	}

	for id, count := range cardList {
		card := db.GetCardByID(id)
		fmt.Printf("%d %s (%s) %s\n", count, card.Name, card.Set, card.CollectorNumber)
	}

	if *inventory {
		f.Seek(0, 0) // Rewind file.
		inventories, err := collectionfinder.FindInventory(f)
		if err != nil {
			log.Fatalf("failed to parse mtga logs: %v", err)
		}
		if len(inventories) == 0 {
			log.Fatalf("No inventory found")
		}
		fmt.Printf("%+v\n", inventories[0])
	}
}
