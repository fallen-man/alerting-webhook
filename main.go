package main

import (
	"bytes"
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

const targetUrl = "http://api.multimedia.mofid.dc/api/Sender/sendsms"

type Config struct {
	Keys []string `yaml:"keys"`
}

type Payload struct {
	ReportFilter ReportFilter `json:"reportFilter"`
}

type ReportFilter struct {
	SenderApplicationCode string   `json:"senderApplicationCode"`
	MediaType             string   `json:"mediaType"`
	Body                  string   `json:"body"`
	Recipients            []string `json:"recipients"`
}

type Response struct {
	Result    string `json:"result"`
	Message   string `json:"message"`
	ErrorFlag bool   `json:"errorFlag"`
}

func main() {
	go getConfig()
	requestHandle()
}

// start a mux router instance for handling incoming requests
func requestHandle() {
	Router := mux.NewRouter().StrictSlash(true)
	Router.HandleFunc("/getMessages", getMessages)
	http.ListenAndServe("localhost:7777", Router)
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
	err := sendRequest(targetUrl, strResult)
	if err != nil {
		log.Warn("error sending request: ", err)
	}
}

// append all input json key:value and ignore configuration file
func appenAll(m map[string]interface{}) {
	keys := sortKeys(m)
	var strResult string
	for _, k := range keys {
		strResult = fmt.Sprintf("%v\n%v: %v", strResult, k, m[k])
	}
	fmt.Println(strResult) // must replace with sendMessage() function
	err := sendRequest(targetUrl, strResult)
	if err != nil {
		log.Warn("error sending request: ", err)
	}
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
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		configBypass = true
		log.Warn("config.yml file not found, application will not filter any data!")
	} else {
		log.Info("config.yml file found.")
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
			log.Info("bypass-filter key found, application will not filter any data!")
		}
	}
}

// initialize payload structure and send the request to target webhook
func sendRequest(url string, payloadData string) error {
	payload := &Payload{
		ReportFilter: ReportFilter{
			SenderApplicationCode: "15",
			MediaType:             "PublicSms",
			Body:                  payloadData,
			Recipients:            []string{"09392922123"},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else {
		checkResponse(resp)
	}
	defer resp.Body.Close()
	return nil
}

func checkResponse(r *http.Response) {
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Warn(err)
	}
	var response Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Warn(err)
	}
	resResult := fmt.Sprintf("result: %v\nmessage: %v\nerror flag: %v", response.Result, response.Message, response.ErrorFlag)
	log.Info("Send Result:\n", resResult)
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
