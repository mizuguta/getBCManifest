package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(fmt.Sprintf("envfiles/%s.env", os.Getenv("GO_ENV")))
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	account_id = os.Getenv("account_id")
	policy_key = os.Getenv("policy_key")
	var configStr = os.Getenv("configs")
	if configStr != "" {
		addConfigIds(configStr)
	}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/getManifest", ManifestIndex)
	router.HandleFunc("/getManifest/{videoId}/", GetAllManifestVideo)
	router.HandleFunc("/getManifest/{videoId}/{configName}", GetManifestVideo)

	log.Fatal(http.ListenAndServe(":8081", router))
}

func addConfigIds(envStr string) {
	for _, configStr := range strings.Split(envStr, ",") {
		configPair := strings.Split(configStr, "=")
		configIds[configPair[0]] = "?config_id=" + configPair[1]
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func ManifestIndex(w http.ResponseWriter, r *http.Request) {
	keys := make([]string, 0)
	for cId := range configIds {
		keys = append(keys, cId)
	}
	fmt.Fprintf(w, "Please specifiy: GET %q/{vidoe_id}/{cofigId}\nconfigIds={\n\t%s\n}", html.EscapeString(r.URL.Path), strings.Join(keys, ",\n\t"))
}
func GetAllManifestVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoId := vars["videoId"]
	var allConfingManifests []ManifestsPerConfig

	for configName := range configIds {
		var configManifest ManifestsPerConfig
		configManifest.ConfigName = configName
		configManifest.ConfigId = configIds[configName]
		manifest, err := GetManifestInfo(videoId, configName)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		configManifest.Manifests = manifest
		allConfingManifests = append(allConfingManifests, configManifest)
	}
	outputJson, err := json.Marshal(allConfingManifests)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outputJson))
}

func GetManifestVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoId := vars["videoId"]
	var configName string = "default"
	if _, ok := vars["configName"]; ok {
		configName = vars["configName"]
	}
	manifest, err := GetManifestInfo(videoId, configName)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	outputJson, err := json.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outputJson))
}
