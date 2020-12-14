package main

import (
	"net/http"

	routes "app-mask/router"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	router := routes.NewRouter()
	http.ListenAndServe(":3000", router)

	// http.HandleFunc("/book", book)

	// http.HandleFunc("/hello", hello)
	// http.HandleFunc("/headers", headers)

	// http.ListenAndServe(":8090", nil)
}
