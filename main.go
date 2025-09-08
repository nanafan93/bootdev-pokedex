package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"pokedex/internal/cachingClient"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
}

type LocationArea struct {
	Name string
	Url  string
}
type LocationAreaResponse struct {
	Count    int
	Next     string
	Previous string
	Results  []LocationArea
}

type PokeExplore struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string
		}
	} `json:"pokemon_encounters"`
}

var nextUrl = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
var previousUrl string
var commandRegistry map[string]cliCommand

type Pokemon struct {
	Name           string
	BaseExperience int `json:"base_experience"`
	Height         int
	Weight         int
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string
		}
	}
	Types []struct {
		Type struct {
			Name string
		}
	}
}

var pokemonCaught = make(map[string]Pokemon)

func init() {
	commandRegistry = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    exitPokedex,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    showHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the nextUrl 20 location areas in the Pokedex",
			callback:    showMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 areas in the Pokedex",
			callback:    showMapB,
		},
		"explore": {
			name:        "explore",
			description: "Explore a specific location area by name",
			callback:    exploreLocation,
		},
		"catch": {
			name:        "catch",
			description: "Tries to catch a specific pokemon by name",
			callback:    catchPokemon,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught pokemon by name",
			callback:    inspectPokemon,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List all caught pokemon",
			callback:    listCaughtPokemon,
		},
	}
}

func listCaughtPokemon(args []string) error {
	if argsErr := checkArgLength(args, 0); argsErr != nil {
		return argsErr
	}
	if len(pokemonCaught) == 0 {
		fmt.Println("You haven't caught any pokemon yet!")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for name := range pokemonCaught {
		fmt.Printf("- %s\n", name)
	}
	return nil
}

func inspectPokemon(args []string) error {
	if argsErr := checkArgLength(args, 1); argsErr != nil {
		return argsErr
	}
	if pokemon, exists := pokemonCaught[strings.ToLower(args[0])]; exists {
		fmt.Printf("Name: %s\nBase Experience: %d\nHeight: %d\nWeight: %d\nTypes:\n", pokemon.Name, pokemon.BaseExperience, pokemon.Height, pokemon.Weight)
		for _, t := range pokemon.Types {
			fmt.Printf("- %s\n", t.Type.Name)
		}
		fmt.Printf("Stats:\n")
		for _, s := range pokemon.Stats {
			fmt.Printf("- %s: %d\n", s.Stat.Name, s.BaseStat)
		}
		return nil
	} else {
		return fmt.Errorf("you haven't caught %s yet", args[0])
	}
}

func catchPokemon(args []string) error {
	if argsErr := checkArgLength(args, 1); argsErr != nil {
		return argsErr
	}
	if _, exists := pokemonCaught[strings.ToLower(args[0])]; exists {
		return fmt.Errorf("you already caught %s", args[0])
	}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", strings.ToLower(args[0]))
	var pokemon = Pokemon{}
	if body, err := cachingClient.Get(url); err != nil {
		return err
	} else {
		fmt.Printf("Throwing a Pokeball at %s...\n", args[0])
		json.Unmarshal(body, &pokemon)
		catchPower := rand.Intn(400)
		if catchPower >= pokemon.BaseExperience {
			pokemonCaught[pokemon.Name] = pokemon
			fmt.Printf("%s was caught!\n", pokemon.Name)
		} else {
			fmt.Printf("%s escaped!\n", pokemon.Name)
		}
		return nil
	}
}

func exploreLocation(args []string) error {
	if argsErr := checkArgLength(args, 1); argsErr != nil {
		return argsErr
	}
	var PokeExp = PokeExplore{}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", strings.ToLower(args[0]))
	if body, err := cachingClient.CachedGet(url); err == nil {
		json.Unmarshal(body, &PokeExp)
		for _, encounter := range PokeExp.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}
		return nil
	} else {
		return err
	}
}

func showMapB(args []string) error {
	if argsErr := checkArgLength(args, 0); argsErr != nil {
		return argsErr
	}
	var currentUrl string
	var loc = LocationAreaResponse{}
	if previousUrl != "" {
		fmt.Println("Going to previous page...")
		currentUrl = previousUrl
	} else {
		fmt.Println("No previous page available.")
		return nil
	}
	if body, err := cachingClient.CachedGet(currentUrl); err != nil {
		return err
	} else {
		json.Unmarshal(body, &loc)
		for _, location := range loc.Results {
			fmt.Println(location.Name)
		}
		nextUrl = loc.Next
		previousUrl = loc.Previous
		fmt.Printf("Next: %v\nPrev: %v\n", nextUrl, previousUrl)
	}
	return nil
}

func showMap(args []string) error {
	if argsErr := checkArgLength(args, 0); argsErr != nil {
		return argsErr
	}
	if nextUrl != "" {
		fmt.Println("Going to nextUrl page...")
	}
	var loc = LocationAreaResponse{}
	if body, err := cachingClient.CachedGet(nextUrl); err == nil {
		json.Unmarshal(body, &loc)
		for _, location := range loc.Results {
			fmt.Println(location.Name)
		}
		nextUrl = loc.Next
		previousUrl = loc.Previous
		fmt.Printf("Next: %v\nPrev: %v\n", nextUrl, previousUrl)
		return nil
	} else {
		return err
	}
}

func showHelp(args []string) error {
	if argsErr := checkArgLength(args, 0); argsErr != nil {
		return argsErr
	}
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\n")
	for name, command := range commandRegistry {
		fmt.Printf("%v: %v\n", name, command.description)
	}
	return nil
}

func exitPokedex(args []string) error {
	if argsErr := checkArgLength(args, 0); argsErr != nil {
		return argsErr
	}
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Pokedex > ")
		if ok := scanner.Scan(); !ok {
			log.Fatal("Error reading input from stdin")
		}
		line := scanner.Text()
		split := cleanInput(line)
		if len(split) == 0 {
			continue
		}
		command := strings.ToLower(split[0])
		args := split[1:]
		tryExecuteCommand(command, args, commandRegistry)
	}
}

func tryExecuteCommand(command string, args []string, registry map[string]cliCommand) {
	if cmd, exists := registry[command]; exists {
		if err := cmd.callback(args); err != nil {
			fmt.Printf("Error executing command : '%s'\n%v\n", command, err)
		}
	} else {
		fmt.Printf("Unknown command: '%s'\n", command)
	}
}

func cleanInput(s string) []string {
	return strings.Fields(s)
}

func checkArgLength(args []string, expectedLength int) error {
	if len(args) != expectedLength {
		return fmt.Errorf("expected %d arguments, got %d", expectedLength, len(args))
	}
	return nil
}
