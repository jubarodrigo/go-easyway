package main

import (
	"template/cmd"
)

// @title Starship Enterprise API
// @version 1.0
// @description This is the Starship Enterprise server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.starshipenterprise.com
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host amazonaws.com/dev
// @BasePath /dev

func main() {
	cmd.StartApp()
}
