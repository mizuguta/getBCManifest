package main

type Rendition struct {
	Bandwidth string `json:"bandwidth"`
	Width     string `json:"width"`
	Height    string `json:"height"`
}
type ManifestInfo struct {
	Url        string      `json:"url"`
	Cdn        string      `json:"cdn"`
	Renditions []Rendition `json:"renditions"`
}

type Manifests map[string]ManifestInfo

type ManifestsPerConfig struct {
	ConfigName string     `json:"config_name"`
	ConfigId   string     `json:"config_id"`
	Manifests  *Manifests `json:"manifests"`
}
