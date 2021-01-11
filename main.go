package main

import (
	"app-mask/controller"
	routes "app-mask/router"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ip, ok := os.LookupEnv("DB_ADDRESS")
	if !ok {
		log.Println("DB_ADDRESS not set.")
		return
	}
	log.Printf("DB_ADDRESS is %s\n", ip)
	controller.InitDBAddress(ip)

	router := routes.NewRouter()
	// http.ListenAndServe(":3000", router)

	// router := mux.NewRouter()
	httpsRouter := routes.RedirectToHTTPSRouter(router)
	// log.Fatal(http.ListenAndServe(lib.Settings.Address, httpsRouter))
	http.ListenAndServe(":3000", httpsRouter)
}
