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
	"strconv"
)

var BASE_URL string
var USERNAME string
var PASSWORD string

var PROBABILITY_SYSTEM, _ = strconv.ParseFloat(os.Getenv("TERMINUSDB_PROB_SYSTEM"), 64)
var PROBABILITY_DB, _ = strconv.ParseFloat(os.Getenv("TERMINUSDB_PROB_DB"), 64)
var PROBABILITY_REPO, _ = strconv.ParseFloat(os.Getenv("TERMINUSDB_PROB_REPO"), 64)
var PROBABILITY_BRANCH, _ = strconv.ParseFloat(os.Getenv("TERMINUSDB_PROB_BRANCH"), 64)

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
		fmt.Printf(`{"severity": "DEBUG", "message": "%s", "path": "%s"}`+"\n", message, path)
	} else {
		message := fmt.Sprintf("Optimize %s failed", path)
		fmt.Printf(`{"severity": "ERROR", "message": "%s", "path": "%s"}`+"\n", message, path)
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
	// This should be filtered by fluentd already
	if logEntry.DescriptorAction != "commit" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if logEntry.Descriptor.DescriptorType == "system" {
		optimizeSystem()
	}
	if logEntry.Descriptor.Branch != "" {
		optimizeBranch(&logEntry.Descriptor)
	}
	if logEntry.Descriptor.Database != "" {
		optimizeDatabase(&logEntry.Descriptor)
	}
	if logEntry.Descriptor.Repository != "" {
		optimizeRepo(&logEntry.Descriptor)
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
