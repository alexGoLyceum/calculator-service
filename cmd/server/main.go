package main

import "github.com/alexGoLyceum/calculator-service/internal/application"

func main() {
	app := application.New()
	app.RunServer()
}
