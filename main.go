package main

import "log"

func main() {
	if err := app(); err != nil {
		log.Fatalln(err)
	}
}

func app() error {
	return nil
}
