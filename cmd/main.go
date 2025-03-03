package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"user_db_test/config"
	scheduler2 "user_db_test/internal/handlers"
	"user_db_test/logger"
)

func main() {
	conf := config.ReadConfig()
	newLogger := logger.NewLogger("taskScheduler.log", "INFO")

	scheduler := scheduler2.NewJobScheduler(newLogger)
	router := mux.NewRouter()
	scheduler.SetupRoutes(router)

	fmt.Printf("Service started on port %v", conf.RESTApi.Port)
	httpPort := ":" + strconv.Itoa(conf.RESTApi.Port)
	http.ListenAndServe(httpPort, router)
}
