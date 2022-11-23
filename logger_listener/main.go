package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
)

var BASE_URL string
var USERNAME string
var PASSWORD string

const PROBABILITY_SYSTEM = 0.1
const PROBABILITY_DB = 0.1
const PROBABILITY_REPO = 0.1
const PROBABILITY_BRANCH = 0.1

type LogEntry struct {
	Descriptor       Descriptor `json:"descriptor"`
	DescriptorAction string     `json:"descriptorAction"`
}

type Descriptor struct {
	DescriptorType string `json:"descriptorType"`
	Organization   string `json:"organization"`
	Repository     string `json:"repository"`
	Database       string `json:"database"`
	Branch         string `json:"branch"`
}

func sendOptimize(path string) {
	client := &http.Client{}
	Url := BASE_URL + url.PathEscape(path)
	req, _ := http.NewRequest("POST", Url, nil)
	req.SetBasicAuth(USERNAME, PASSWORD)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode == 200 {
		message := fmt.Sprintf("Optimize %s completed succesfully", path)
		fmt.Printf(`{"sevirity": "DEBUG", "message": "%s", "path": "%s"}` + "\n", message, path)
	} else {
		message := fmt.Sprintf("Optimize %s failed", path)
		fmt.Printf(`{"sevirity": "ERROR", "message": "%s", "path": "%s"}` + "\n", message, path)
	}
}

func optimizeSystem() {
	if rand.Float64() <= PROBABILITY_SYSTEM {
		sendOptimize("_system")
	}
}

func optimizeBranch(descriptor *Descriptor) {
	if rand.Float64() <= PROBABILITY_BRANCH {
		path := fmt.Sprintf("%s/%s/%s/branch/%s", descriptor.Organization, descriptor.Database, descriptor.Repository, descriptor.Branch)
		sendOptimize(path)
	}
}

func optimizeRepo(descriptor *Descriptor) {
	if rand.Float64() <= PROBABILITY_REPO {
		path := fmt.Sprintf("%s/%s/%s/_commits", descriptor.Organization, descriptor.Database, descriptor.Repository)
		sendOptimize(path)
	}
}

func optimizeDatabase(descriptor *Descriptor) {
	if rand.Float64() <= PROBABILITY_DB {
		path := fmt.Sprintf("%s/%s/_meta", descriptor.Organization, descriptor.Database)
		sendOptimize(path)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	raw_message, _ := ioutil.ReadAll(r.Body)
	if !json.Valid(raw_message) {
		w.WriteHeader(http.StatusOK)
		return
	}
	var logEntry *LogEntry
	err := json.Unmarshal(raw_message, &logEntry)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	// TODO: This should be filtered by fluentd already
	if logEntry.DescriptorAction != "commit" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if logEntry.Descriptor.DescriptorType == "system" {
		optimizeSystem()
	} else if logEntry.Descriptor.Database != "" {
		optimizeDatabase(&logEntry.Descriptor)
	} else if logEntry.Descriptor.Repository != "" {
		optimizeRepo(&logEntry.Descriptor)
	} else if logEntry.Descriptor.Branch != "" {
		optimizeBranch(&logEntry.Descriptor)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	fmt.Println("STARTING SERVER")
	BASE_URL = os.Getenv("TERMINUSDB_BASE_HOST") + "/api/optimize/"
	USERNAME = os.Getenv("TERMINUSDB_USERNAME")
	PASSWORD = os.Getenv("TERMINUSDB_PASSWORD")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
