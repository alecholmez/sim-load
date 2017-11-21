package main

type service struct {
	Name     string   `toml:"name" json:"name"`
	Routes   []string `toml:"routes" json:"routes"`
	Location string   `toml:"location" json:"location"`
	Load     string   `toml:"load" json:"load"`
}

type services struct {
	Service []service `json:"services"`
}

type key int
