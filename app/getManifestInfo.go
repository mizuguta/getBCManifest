package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var account_id string
var policy_key string
var playback_url = "https://edge.api.brightcove.com/playback/v1/accounts/%s/videos/%s"

var configIds map[string]string = map[string]string{
	"default": "",
}

type VideoObject struct {
	AccountId string `json:"account_id"`
	Sources   []struct {
		Type          string `json:"type"`
		Ext_x_version string `json:"ext_x_version"`
		Src           string `json:"src"`
		Profiles      string `json:"profiles"`
	} `json:"sources"`
}

func GetManifestInfo(videoId string, configName string) (*Manifests, error) {
	manifests := make(Manifests)

	videoObj, err := getVideo(videoId, configName)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(videoObj.Sources); i++ {
		if videoObj.Sources[i].Type == "application/x-mpegURL" {
			if _, ok := manifests["hls"]; !ok {
				manifests["hls"] = dumpHls(videoObj.Sources[i].Src)
			}
		}
		if videoObj.Sources[i].Type == "application/dash+xml" {
			if _, ok := manifests["dash"]; !ok {
				manifests["dash"] = dumpDash(videoObj.Sources[i].Src)
			}
		}
		if videoObj.Sources[i].Type == "application/vnd.ms-sstr+xml" {
			if _, ok := manifests["smooth"]; !ok {
				manifests["smooth"] = dumpSmooth(videoObj.Sources[i].Src)
			}
		}
	}
	return &manifests, nil
}

func getBaseFQDN(url string) string {
	scheme := ""
	if strings.Contains(url, "https://") {
		scheme = "https://"
	}
	if strings.Contains(url, "http://") {
		scheme = "http://"
	}
	if scheme == "" {
		return ""
	}
	firstInx := strings.Index(url, scheme) + len(scheme)
	lastInx := strings.Index(url[firstInx:], "/")
	baseUrl := url[firstInx : lastInx+firstInx]
	return baseUrl
}

func dumpHls(url string) ManifestInfo {
	hlsData := ManifestInfo{}
	hlsData.Url = url

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("HLS manigest error")
		return hlsData
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "URI=") {
			subUrl := line[strings.Index(line, "URI=")+5 : len(line)-1]
			subresp, err := http.Get(subUrl)
			if err != nil {
				fmt.Println("HLS manigest error")
				return hlsData
			}
			defer subresp.Body.Close()

			subManiData, err := ioutil.ReadAll(subresp.Body)
			if err != nil {
				log.Fatal(err)
				return hlsData
			}
			hlsData.Cdn = getBaseFQDN(string(subManiData))
		}
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			var bandwidth, width, height string
			elems := strings.Split(line, ",")
			for _, elem := range elems {
				if strings.HasPrefix(elem, "BANDWIDTH=") {
					bandwidth = strings.Trim(elem, "BANDWIDTH=")
				}
				if strings.HasPrefix(elem, "RESOLUTION=") {
					res := strings.Trim(elem, "RESOLUTION=")
					wxh := strings.Split(res, "x")
					width = wxh[0]
					height = wxh[1]
				}
			}
			hlsData.Renditions = append(hlsData.Renditions, Rendition{bandwidth, width, height})
		}
	}
	return hlsData
}

func dumpDash(url string) ManifestInfo {
	dashData := ManifestInfo{}
	dashData.Url = url

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Dash manigest error")
		return dashData
	}
	defer resp.Body.Close()

	xmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return dashData
	}
	type Representation struct {
		Bandwidth string `xml:"bandwidth,attr"`
		Width     string `xml:"width,attr"`
		Height    string `xml:"height,attr"`
	}
	type DashManifest struct {
		BaseUrl         string           `xml:"BaseURL"`
		Representations []Representation `xml:"Period>AdaptationSet>Representation"`
	}
	var dashManifest DashManifest
	xml.Unmarshal(xmlData, &dashManifest)
	dashData.Cdn = getBaseFQDN(dashManifest.BaseUrl)
	for _, seg := range dashManifest.Representations {
		if seg.Width != "" {
			dashData.Renditions = append(dashData.Renditions, Rendition{seg.Bandwidth, seg.Width, seg.Height})
		}
	}
	return dashData
}

func dumpSmooth(url string) ManifestInfo {
	smoothData := ManifestInfo{}
	smoothData.Url = url

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Smooth manigest error")
		return smoothData
	}
	defer resp.Body.Close()

	xmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return smoothData
	}
	type QualityLevel struct {
		Bandwidth string `xml:"Bitrate,attr"`
		Width     string `xml:"MaxWidth,attr"`
		Height    string `xml:"MaxHeight,attr"`
	}
	type SmoothManifest struct {
		QualityLevels []QualityLevel `xml:"StreamIndex>QualityLevel"`
	}
	var smoothManifest SmoothManifest
	xml.Unmarshal(xmlData, &smoothManifest)
	smoothData.Cdn = getBaseFQDN(url)
	for _, seg := range smoothManifest.QualityLevels {
		if seg.Width != "" {
			smoothData.Renditions = append(smoothData.Renditions, Rendition{seg.Bandwidth, seg.Width, seg.Height})
		}
	}
	return smoothData
}

func getVideo(video_id string, configName string) (*VideoObject, error) {
	fmt.Println("======= playback information =======")
	url := fmt.Sprintf(playback_url, account_id, video_id) + configIds[configName]
	fmt.Println(url)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "BCOV-Policy "+policy_key)
	req.Header.Set("Content-Type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("get video object error!")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		errorMsg, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s %s", resp.Status, string(errorMsg))
	}
	var videoObj = new(VideoObject)

	jsonErr := json.NewDecoder(resp.Body).Decode(videoObj)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return nil, nil
	}
	return videoObj, nil
}
