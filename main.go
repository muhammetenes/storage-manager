package main

import (
	"fmt"
	"main/router"
)

func main() {
	e := router.New()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", router.Port)))
}
