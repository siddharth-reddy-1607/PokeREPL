package PokeAPI
import (
        "fmt"
        "errors"
    )
    

var validAPIs = map[string]APIRequest{"Location Area": {EndPoint:"https://pokeapi.co/api/v2/location-area/",
                                                         QueryParams:map[string]int{"offset":0,
                                                                                    "limit":20}},
                                      "Pokemons": {EndPoint:"https://pokeapi.co/api/v2/pokemon/",
                                                   QueryParams:make(map[string]int)}}
type APIRequest struct{
    EndPoint string
    QueryParams map[string]int
}

func Get_API_endpoint(name string) (*APIRequest,error){
    api,ok := validAPIs[name]
    if (!ok){
        return &APIRequest{},errors.New(fmt.Sprintf("%s is not a valid API Name",name))
    }
    return &api,nil
}
