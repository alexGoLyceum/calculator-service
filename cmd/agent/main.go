package main

import "github.com/alexGoLyceum/calculator-service/agent/app"

func main() {
	application := app.NewApplication()
	application.Start()
}
