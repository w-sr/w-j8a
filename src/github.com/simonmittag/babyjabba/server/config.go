package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

//Server config of all Routes mapped to Resources
type ServerConfig struct {
	Routes    []Route
	Resources []Resource
}

//special Route alias for internal endpoint
const AboutJabba string = "aboutJabba"

//Route maps a Path to an upstream resource
type Route struct {
	Path  string
	Alias string
	Label string
}

//Resource describes upstream servers
type Resource struct {
	Alias    string
	Labels   []string
	Upstream Upstream
}

//Upstream describes host mapping
type Upstream struct {
	Scheme string
	Host   string
	Port   int16
}

//String representation of our URL struct
func (u Upstream) String() string {
	return u.Scheme + "://" + u.Host + ":" + strconv.Itoa(int(u.Port))
}

//ServerConfig stores global params
var Live ServerConfig

func parseFromFile() *ServerConfig {
	jsonFile, err := os.Open("babyjabba.json")
	defer jsonFile.Close()
	if err != nil {
		msg := "cannot find babyjabba.json, unable to read server configuration, exiting..."
		log.Fatal().Msg(msg)
		panic(msg)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &Live)
	//todo tell me about more of the config, number of routes
	log.Debug().Msgf("parsed server configuration with %d live routes", len(Live.Routes))
	return &Live
}
