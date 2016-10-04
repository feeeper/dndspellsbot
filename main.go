package main

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"encoding/json"
	"encoding/xml"
	"strings"
)

type Config struct {
	TelegramBotToken string
}

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(configuration.TelegramBotToken)

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}

	spells, err:= parseSpells()
	if err != nil {
		log.Panic(err)
	}

	for update := range updates { 
		query := update.Message.Text
		filteredSpells := Filter(spells.Spells, func(spell Spell) bool { 
			return strings.Index(strings.ToLower(spell.Name), strings.ToLower(query)) >= 0
		})

		if len(filteredSpells) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "No one spells matches")
			bot.Send(msg)			
		}

		for _, spell := range(filteredSpells) {
			text := ""
			for _, t := range(spell.Texts) {
				text = text + t + "\n"
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s\n%s", spell.Name, text))
			bot.Send(msg)
		}
	}
}

type Spells struct {
	XMLName xml.Name `xml:"compendium"`
	Spells []Spell `xml:"spell"`
}

type Spell struct {
	XMLName xml.Name `xml:"spell"`
	Name string `xml:"name"`
	Level int `xml:"level"`
	School string `xml:"school"`
	Time string `xml:"time"`
	Range string `xml:"range"`
	Components string `xml:"components"`
	Duration string `xml:"duration"`
	Classes string `xml:"classes"`
	Texts []string `xml:"text"`
	Rolls []string `xml:"roll"`
}

func parseSpells() (Spells, error) {
	file, err := os.Open("phb.xml")
	if err != nil {
		log.Panic(err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Panic(err)
	}

	var data = make([]byte, fi.Size())
	_, err = file.Read(data)
	if err != nil {
		log.Panic(err)
	}

	var v Spells
	err = xml.Unmarshal(data, &v)
	
	if err != nil {
		log.Println(err)
		return v, err
	} else {
		log.Printf("Total spells found: %v", len(v.Spells))
		return v, err
	}
}

func Filter(spells []Spell, fn func(spell Spell) bool) []Spell {
	var filtered []Spell
	for _, spell := range(spells) {
		if fn(spell) {
			filtered = append(filtered, spell)
		}
	}
	return filtered
}