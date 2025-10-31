package main

import (
	"flag"
	"log"
)

type Serominers struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	State     string `json:"state"`
	CreatedOn string `json:"created_on"`
}

/*
 * Create a new device.
 • Fully and/or partially update an existing device.
 • Fetch a single device.
 • Fetch all devices.
 • Fetch devices by brand.
 • Fetch devices by state.
 • Delete a single device.
*/

func main() {
	var httpPort int
	var httpHost string
	var err error

	var app ApiApplication = ApiApplication{}

	flag.IntVar(&httpPort, "port", 9098, "The port on which the API server listens")
	flag.StringVar(&httpHost, "host", "0.0.0.0", "The host on which the API server listens")
	flag.Parse()

	if err = app.Setup(httpHost, httpPort); err != nil {
		log.Println("Error: ", err)
	}

	log.Fatal(app.Run())
}
