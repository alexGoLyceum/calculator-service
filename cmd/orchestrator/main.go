package main

import "github.com/alexGoLyceum/calculator-service/orchestrator/app"

func main() {
	application := app.NewApplication()
	application.Start()
}
