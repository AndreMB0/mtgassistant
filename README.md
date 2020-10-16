# MTG Arena Assistant
Helper code for Magic The Gathering: Arena

# Requirements
You will need a computer that can run `Magic The Gathering: Arena`, and set it up so it exports logs.
To do so, open the game, go to Settings > View Account > Enable Logs, then restart the game and exit.

These programs work by parsing the resource files inside your installation of MTG: Arena, as well as
the game logs. If those things are not in the standard locations, you will need to specify those to
the programs via command-line flags.

# What can I do?
Right now the assistant only has two binaries: a collection exporter, and a deck helper.

## MTGA Assistant
MTGA Assistant is a web application that parse your MTGA log and print the cards you have found in the latest boosters opening.
It tracks the wildcard you found and can redeem your wildcards with a related valid card in the current set.

How to use:
- Download the release package
- Using the Terminal or Command Prompt, go to the downloaded folder
- Execute the following command (please use /usr/local/go/bin/go if you have problems finding the go command) :
```
$ go run main.go
```
- If you don't have the MTGA client Data folder (C:\Program Files\Wizards of the Coast\MTGA\MTGA_Data\Downloads\Data), you can use the flag --mtg_data for pointing to a directory containing the data files, example:
```
/usr/local/go/bin/go run main.go --mtg_data "../Data"
```
- You can always set a different Set for wildcard picking using the flag --set, example:
```
/usr/local/go/bin/go run main.go --set "IKO"
```

## Collection Exporter
Collection Exporter is a program that will parse your MTG:A collection and print it in the MTG:A format.
To run it, just do:

```
$ go run collectionexporter/main.go
```

## Deck Helper
Given a decklist in the MTG:A format, deck helper will parse your MTG:A collection and will tell you
which cards are missing from the decks (cards that you need to craft). Currently it only supports the
standard rotation, but that can be easily changed from the code.

To run it:

```
$ go run deckhelper.go -deck=<path-to-your-deck>
```

# Libraries

There's a `carddb` library that parses the resource files and creates a database of magic cards. You can
use it to make queries based on card names, card ids, or just iterate over it and run the code that you
want.

There's also a `collectionfinder` library that parses the game logs and gets your card collection.
