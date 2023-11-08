package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
func requestHandle() {
	Router := mux.NewRouter().StrictSlash(true)
	Router.HandleFunc("/getMessages", getMessages)
	http.ListenAndServe("localhost:9090", Router)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("error reading body:", err)
		return
	}
	defer r.Body.Close()
	jsonParse(bodyData)
}

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

func customAppend(m map[string]interface{}) {
	var strResult string
	for key, value := range m {
		for _, configKey := range config.Keys {
			if key == configKey {
				strResult = fmt.Sprintf("%v\n%v: %v", strResult, key, fmt.Sprint(value))
			}
		}
	}
	fmt.Println(strResult)
}

func appenAll(m map[string]interface{}) {
	var strResult string
	for key, value := range m {
		strResult = fmt.Sprintf("%v\n%v: %v", strResult, key, fmt.Sprint(value))
	}
	fmt.Println(strResult)
}

func getConfig() {
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("error: ", err)
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatal("error: ", err)
	}
	if config.Keys[0] == "BYPASS-FILTER" {
		configBypass = true
	}
}
