package main

import (
	"api/cmd"
	"api/docs"
	"os"
)

// @title			UWPSC API
// @version		1.0
// @description	This is the API for the UWPSC website.
//
// @contact.name	UWPSC Development Team
// @contact.email	uwaterloopoker@gmail.com
//
// @license.name	Apache 2.0
// @license.url	https://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "localhost:5000/api"
	}

	// Set Swagger info
	docs.SwaggerInfo.Host = hostname
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	cmd.Execute()
}
