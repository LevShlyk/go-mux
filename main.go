package main

import "os"

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("DB"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DBNAME"),
	)

	a.Run(":8010")
}
