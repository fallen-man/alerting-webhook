package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	requestHandle()
}
func requestHandle() {
	//creating new mux router for handling requests
	Router := mux.NewRouter().StrictSlash(true)
	// with mux vars can be passed with placeholders {id} is a var
	Router.HandleFunc("/getMessages", getMessages)
	http.ListenAndServe("localhost:9090", Router)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(os.Stdout).Encode(r.Header)
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("error reading body:", err)
		return
	}

	defer r.Body.Close()

	var result map[string]interface{}
	err = json.Unmarshal(bodyData, &result)
	if err != nil {
		log.Fatal("Error parsing Json:", err)
		return
	}

	fmt.Println(result)
}
