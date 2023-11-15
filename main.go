package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Global Variables
var config Config
var configBypass, aofBypass bool

const targetURL = "http://api.multimedia.mofid.dc/api/Sender/sendsms"

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
	startHTTPServer()
}

// startHTTPServer initializes a mux router and starts an HTTP server
func startHTTPServer() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/getMessages", getMessages)
	server := &http.Server{
		Addr:         "0.0.0.0:7777",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Info("HTTP server listening on localhost:7777")
	log.Fatal(server.ListenAndServe())
}

// getMessages handles requests to the /getMessages endpoint
func getMessages(w http.ResponseWriter, r *http.Request) {
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("Error reading request body:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	jsonParse(bodyData)
	if !aofBypass {
		go saveRawData(bodyData)
	}

}

// jsonParse parses the JSON body of an incoming request
func jsonParse(b []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(b, &result)
	if err != nil {
		log.Error("Error parsing JSON:", err)
		return make(map[string]interface{}), err
	}
	if !configBypass {
		customAppend(result)
	} else {
		appendAll(result)
	}

	return result, nil
}

// customAppend appends input JSON key:value pairs based on the configuration file
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
	log.Println(strResult) // Replace with sendMessage() function
	err := sendRequest(targetURL, strResult)
	if err != nil {
		log.Warn("Error sending request:", err)
	}
}

// appendAll appends all input JSON key:value pairs and ignores the configuration file
func appendAll(m map[string]interface{}) {
	keys := sortKeys(m)
	var strResult string
	for _, k := range keys {
		strResult = fmt.Sprintf("%v\n%v: %v", strResult, k, m[k])
	}
	log.Println(strResult) // Replace with sendMessage() function
	err := sendRequest(targetURL, strResult)
	if err != nil {
		log.Warn("Error sending request:", err)
	}
}

// sortKeys creates a slice from a map and sorts the content of the slice
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

// getConfig parses YAML configuration file content and sets conditions for parsing JSON
func getConfig() {
	const configFileName = "config.yml"
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		configBypass = true
		log.Warnf("%s file not found, application will not filter any data!", configFileName)
	} else {
		configData, err := os.ReadFile(configFileName)
		if err != nil {
			log.Fatal("Error reading config file:", err)
		}
		err = yaml.Unmarshal(configData, &config)
		if err != nil {
			log.Fatal("Error parsing config file:", err)
		}
		if contains(config.Keys, "bypass-filter") {
			configBypass = true
			log.Info("Bypass-filter key found, application will not filter any data!")
		}
		if contains(config.Keys, "bypass-aof") {
			aofBypass = true
			log.Info("aof-bypass key found, application will not save raw alerts data in append-only file")
		}
	}
}

// sendRequest initializes payload structure and sends the request to the target webhook
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
		return fmt.Errorf("Error marshaling payload: %v", err)
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()
	checkResponse(resp)
	return nil
}

// checkResponse reads and logs the response from the HTTP request
func checkResponse(r *http.Response) {
	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Warn("Error reading response body:", err)
	}
	var response Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		log.Warn("Error parsing response JSON:", err)
	}
	resResult := fmt.Sprintf("Result: %v\nMessage: %v\nError Flag: %v", response.Result, response.Message, response.ErrorFlag)
	log.Info("Send Result:\n", resResult)
}

// saveRawData will save raw json in file on localmachine
func saveRawData(b []byte) {
	dir := "/etc/alerting-webhook"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			log.Errorf("Error creating data directory: %v", err)
		}
	}

	time := time.Now()
	fileName := filepath.Join(dir, "raw-alerts-"+time.Format("2023-01-02")+".json")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(b); err != nil {
		log.Errorf("Error writing to file: %v", err)
	}
}

// contains checks if a string is present in a slice
func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
