package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	yaml "gopkg.in/yaml.v3"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Global Variables
var config Config
var configBypass bool

type Config struct {
	Keys []string `yaml:"keys"`
}

func main() {
	go getConfig()
	requestHandle()
}

// start a mux router instance for handling incoming requests
func requestHandle() {
	Router := mux.NewRouter().StrictSlash(true)
	Router.HandleFunc("/getMessages", getMessages)
	http.ListenAndServe("localhost:9090", Router)
}

// handle every request that hit the /getMessages endpoint
func getMessages(w http.ResponseWriter, r *http.Request) {
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("error reading body:", err)
		return
	}
	defer r.Body.Close()
	jsonParse(bodyData)
}

// parse the json body of incoming request
func jsonParse(b []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(b, &result)
	if err != nil {
		log.Fatal("Error parsing Json:", err)
		return make(map[string]interface{}, 0), err
	}
	if configBypass != true {
		customAppend(result)
	} else {
		appenAll(result)
	}

	return result, nil
}

// append input json key:value based on configuration file
func customAppend(m map[string]interface{}) {
	keys := sortKeys(m)
	var strResult string
	for _, k := range keys {
		for _, configKey := range config.Keys {
			if k == configKey {
				strResult = fmt.Sprintf("%v\n%v: %v", strResult, k, m[k])
			}
		}
	}
	fmt.Println(strResult) // must replace with sendMessage() function
}

// append all input json key:value and ignore configuration file
func appenAll(m map[string]interface{}) {
	keys := sortKeys(m)
	var strResult string
	for _, k := range keys {
		strResult = fmt.Sprintf("%v\n%v: %v", strResult, k, m[k])
	}
	fmt.Println(strResult) // must replace with sendMessage() function
}

// create a slice from map and sort the content of the slice
func sortKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// parse YAML configuration file content and set the conditions of parsing json
func getConfig() {
	configData, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatal("error: ", err)
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatal("error: ", err)
	}
	if contains(config.Keys, "bypass-filter") {
		configBypass = true
	}
}

// implementing "in" function of python for finding values in a slice
func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
