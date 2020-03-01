package main

import (
	"fmt"
	"main/config"
	"main/router"
)

func main() {
	e := router.New()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.Port)))
}
