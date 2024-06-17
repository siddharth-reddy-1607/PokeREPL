package main

import (
	"PokeREPL/PokeAPI"
	"PokeREPL/PokeCache"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
)


const EXIT string = "exit"
const HELP string = "help"
const MAP string = "map"
const MAPB string = "mapb"
const EXPLORE string = "explore"
const CATCH string = "catch"
const INSPECT string = "inspect"
const POKEDEX string = "pokedex"

const cacheTimeSeconds = 60


var LocationAreaCacher *pokecache.Cache
var PokemonsInAreaCache *pokecache.Cache
var PokemonCache *pokecache.Cache

type command struct{
    name string
    description string
    api_request *PokeAPI.APIRequest
    exec func(valid_commads map[string]*command, args ...string,) error
}

func CommandExit(valid_commands map[string]*command, args ...string) error{
    os.Exit(0)
    return nil
}

func CommandHelp(valid_commands map[string]*command, args ...string) error{
    if args[0] == ""  {
        fmt.Println("Welcome to the REPL. Here are a list of commands that you can use")
        for command,cmd_struct := range valid_commands{
            fmt.Printf("%s : %s\n",command,cmd_struct.description)
        }
        return nil
    }
    if len(args) > 1{
        return errors.New("Help command cannot take more than 1 argument")
    }
    command_to_help,ok := valid_commands[args[0]]
    if !ok{
        return errors.New(fmt.Sprintf("Command %s not found\n",args[0]))
    }
    fmt.Printf("%s : %s\n",command_to_help.name,command_to_help.description)
    return nil
}

type LocationAreasListerResponse struct{
    Count int `json:"count"`
    Next *string `json:"next"`
    Previous *string `json:"previous"`
    Results []struct{
        Name string `json:"name"`
        URL string `json:"url"`
    } `json:"results"`
}

var location_areas LocationAreasListerResponse

func HandleLocationAreasLister(url string) error{
    body,err := LocationAreaCacher.Get(url)
    if err == nil{
        fmt.Printf("Entry found if cache for %v. Using it.\n",url)
    }
    if err != nil{
        res,err := http.Get(url)
        if err != nil{
            return fmt.Errorf("Error while making a GET request to %s",url)
        }
        defer res.Body.Close()
        body,err = io.ReadAll(res.Body)
        if err != nil {
            return errors.New("Error while reading the body")
        }
        LocationAreaCacher.Add(url,body)
    }
    err = json.Unmarshal(body, &location_areas)
    if err != nil{
        panic(err)
    }
    for _,result := range location_areas.Results{
        fmt.Printf("%s\n",result.Name)
    }
    return nil
}

func CommandMap(valid_commands map[string]*command, args ...string) error{
    if args[0] != ""{
        return errors.New("map command doesn't take any arguments.")
    }
    cmd := valid_commands[MAP]
    if LocationAreaCacher == nil{
        LocationAreaCacher = pokecache.NewCache(cacheTimeSeconds)
    }
    if cmd.api_request == nil{
        api_request,err := PokeAPI.Get_API_endpoint("Location Area")
        if err != nil{
            return err
        }
        cmd.api_request = api_request
        err = HandleLocationAreasLister(cmd.api_request.EndPoint)
        return err
    }
    if location_areas.Next == nil {
        return errors.New("No more locations to show after this. Use map")
    }
    err := HandleLocationAreasLister(*location_areas.Next)
    if err !=nil {
        return err
    }
    return nil
}

func CommandMapBack(valid_commands map[string]*command, args ...string) error{
    if args[0] != "" && len(args) > 1{
        return errors.New("mapb command doesn't take any arguments")
    }
    if LocationAreaCacher == nil{
        LocationAreaCacher = pokecache.NewCache(cacheTimeSeconds)
    }
    cmd := valid_commands[MAP]
    if cmd.api_request == nil{
        return errors.New("No more locations to show before this. Use map")
    }
    if location_areas.Previous == nil {
        return errors.New("No more locations to show before this. Use map")
    }
    err := HandleLocationAreasLister(*location_areas.Previous)
    if err !=nil {
        return err
    }
    return nil
}

type LocationArea struct{
    PokemonEncounters []struct{
        Pokemon struct{
            Name string `json:"name"`
            Url string `json:"url"`
        } `json:"pokemon"`
    } `json:"pokemon_encounters"`
}

var location_area LocationArea

func HandleLocationArea(url string) error{
    body,err := PokemonsInAreaCache.Get(url)
    if err == nil{
        fmt.Printf("Entry found if cache for %v. Using it.\n",url)
    }
    if err!= nil{
        res,err := http.Get(url)
        splitted := strings.Split(url,"/")
        fmt.Printf("Exploring pokemons in location_area %v\n",splitted[len(splitted) -1])
        if err != nil{
            return fmt.Errorf("Error while making a GET Request to %s : %v\n",url,err)
        }
        if res.StatusCode != http.StatusOK{
            return fmt.Errorf("Cannot find pokemons in the location %s due to %v", splitted[len(splitted) - 1],res.Status)
        }
        defer res.Body.Close()
        body,err = io.ReadAll(res.Body)
        if err!= nil{
            return fmt.Errorf("Error while reading from the body: %v",err)
        }
        PokemonsInAreaCache.Add(url,body)
    }
    err = json.Unmarshal(body,&location_area)
    if err != nil{
        panic(err)
    }
    for idx,pokemon := range location_area.PokemonEncounters{
        fmt.Printf("%d : %v\n",idx+1,pokemon.Pokemon.Name)
    }
    return nil 
}

func CommandExplore(valid_commands map[string]*command,args ...string) error{
    if PokemonsInAreaCache == nil{
        PokemonsInAreaCache = pokecache.NewCache(cacheTimeSeconds)
    }
    if args[0] == ""{
        return errors.New("Must pass in location area as an argument to explore")
    }
    if len(args) > 1{
        return errors.New("explore received more that 1 argument. Please pass in the location area you can discover as an argument to the explore command")
    }
    cmd := valid_commands[EXPLORE]
    if cmd.api_request == nil{
        api_request,err := PokeAPI.Get_API_endpoint("Location Area")
        if err != nil{
            return err
        }
        cmd.api_request = api_request
    }
    location_area := args[0]
    url := cmd.api_request.EndPoint + location_area
    err := HandleLocationArea(url)
    if err != nil{
        return err
    }
    return nil
}

type Pokemon struct{
    Name string `json:"name"`
    Height int `json:"height"`
    Weight int `json:"weight"`
    BaseExperience int `json:"base_experience"`
    Types []struct{
        Type struct{
            Name string `json:"name"`
            Url string `json:"url"`
        }`json:"type"`
    }`json:"Types"`
    Stats []struct{
        BaseStat int `json:"base_stat"`
        Effort int `json:"effort"`
        Stat struct{
            Name string `json:"name"`
            Url string `json:"url"`
        }`json:"stat"`
    }`json:"stats"`
}

var pokemon Pokemon
var usersPokemons = make(map[string]*Pokemon)

func handleCatchPokemon(url string) (*Pokemon,error){
    body,err := PokemonCache.Get(url)
    if err == nil{
        fmt.Printf("Entry found if cache for %v. Using it.\n",url)
    }else{
        splitted := strings.Split(url,"/")
        res,err := http.Get(url)
        if err != nil{
            return &Pokemon{},fmt.Errorf("Error while making a GET Request to %s : %v\n",url,err)
        }
        if res.StatusCode != http.StatusOK{
            return &Pokemon{},fmt.Errorf("Cannot find pokemon %s. Please enter a valid pokemon name",splitted[len(splitted) -1])
        }
        body,err = io.ReadAll(res.Body)
        if err!= nil{
            return &Pokemon{},fmt.Errorf("Error while reading from the body: %v",err)
        }
        PokemonCache.Add(url,body)
    }
    err = json.Unmarshal(body,&pokemon)
    if err != nil{
        return &pokemon,err
    }
    return &pokemon,err
}

func CommandCatch(valid_commands map[string]*command, args ...string) error{
    if PokemonCache == nil{
        PokemonCache = pokecache.NewCache(cacheTimeSeconds)
    }
    if args[0] == ""{
        return errors.New("Must pass in pokemon name as an argument to catch")
    }
    if len(args) > 1{
        return errors.New("catch received more that 1 argument. Please pass in the name of the pokemon you want to try and catch as an argument to the catch command")
    }
    pokemon_name := args[0]
    fmt.Printf("Throwing ball at %s...\n",pokemon_name)
    cmd := valid_commands[CATCH]
    if cmd.api_request == nil{
        api_request,err := PokeAPI.Get_API_endpoint("Pokemons")
        if err != nil{
            return err
        }
        cmd.api_request = api_request
    }
    url := cmd.api_request.EndPoint + pokemon_name
    pokemon,err := handleCatchPokemon(url)
    if err != nil{
        return err
    }
    randomNumber := rand.Intn(11) 
    if pokemon.BaseExperience >= 75{
        if randomNumber >= 8{
            fmt.Printf("Yayy!! You caught %s.\nYou can now use the inspect command to view the pokemon details\n",pokemon_name)
            usersPokemons[pokemon_name] = pokemon
        }else{
            fmt.Printf("Missed!! Please try again\n")
        } 
    }else{
        if randomNumber > 4{
            fmt.Printf("Yayy!! You caught %s.\nYou can now use the inspect command to view the pokemon details\n",pokemon_name)
            usersPokemons[pokemon_name] = pokemon
        }else{
            fmt.Printf("Missed!! Please try again\n")
        }
    }
    return nil
}

func CommandInspect(valid_commands map[string]*command,args ...string) error{
    if args[0] == ""{
        return fmt.Errorf("Must pass the name of the pokemon you would like to inspect as argument to inspect")
    }
    if len(args) > 1{
        return fmt.Errorf("inspect received more than 1 argument. Please pass the name of the pokemon you want to inspect")
    }
    pokemon_name := args[0] 
    pokemon_data,ok := usersPokemons[pokemon_name]
    if !ok{
        return fmt.Errorf("You cannot inspect a pokemon you haven't caught. Please catch %s and try again.",pokemon_name)
    }
    fmt.Printf(`Basic Info:
    - Name : %s
    - Height : %d
    - Weight: %d`,pokemon_data.Name,pokemon_data.Height,pokemon_data.Weight)
    fmt.Printf("\nStats\n")
    for _,stat := range pokemon_data.Stats{
        fmt.Printf(" - %d %s for %d effort\n",stat.BaseStat,stat.Stat.Name,stat.Effort)
    }
    fmt.Println("Types")
    for _,t := range pokemon_data.Types{
        fmt.Printf(" - %s\n",t.Type.Name)
    }
    return nil
}

func CommandPokedex(valid_commads map[string]*command, args ...string) error{
    if args[0] != ""{
        return errors.New("pokedex doesn't take any arguments.")
    }
    if len(usersPokemons) == 0{
        fmt.Println("Uhh oh you haven't caught any pokemons yet.\n Use explore command to explore pokemon in a location and catch them using the catch command!")
    }
    for pokemon,_ := range usersPokemons{
        fmt.Println(" -  ",pokemon)
    }
    return nil
}

func main(){
    input_command := ""
    scanner := bufio.NewScanner(os.Stdin)
    valid_commands := map[string]*command{EXIT:{name:"exit",
                                              description: "Exits the REPL",
                                              exec:CommandExit},
                                         HELP:{name:"help",
                                              description: "Describes the usage of the command passed a argument",
                                              exec:CommandHelp},
                                         MAP:{name:"map",
                                              description: "Displays the 20 next location areas in the Pokemon world",
                                              exec:CommandMap},
                                         MAPB:{name:"mapb",
                                               description: "Displays the 20 previous location areas in the Pokemon world",
                                               exec:CommandMapBack},
                                         EXPLORE:{name:"explore",
                                                  description: "Shows the list of pokemons in <location_area_name>",
                                                  exec:CommandExplore},
                                         CATCH:{name:"catch",
                                                description: "Attempts to catch the pokemon <pokemon_name>",
                                                exec:CommandCatch},
                                         INSPECT:{name:"inspect",
                                                  description: "Inspects <pokemon_name> if you caught it",
                                                  exec:CommandInspect},
                                         POKEDEX:{name:"pokedex",
                                                  description: "Displays the names of pokemons you caught so far",
                                                  exec:CommandPokedex}}
    for true{
        fmt.Printf("pokedex > ")
        scanner.Scan()  
        input_command = scanner.Text()
        command_name,args,_ := strings.Cut(input_command," ")
        cmd,ok := valid_commands[command_name]
        if !ok{
           fmt.Printf("Command %s is not a valid command \n",command_name) 
           continue
        }
        err := cmd.exec(valid_commands, strings.Split(args, " ")...)
        if err != nil{
            fmt.Println(err)
        }
    }

}
