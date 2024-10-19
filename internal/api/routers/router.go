package routers

import (
    "github.com/gorilla/mux"
    "net/http"
    "github.com/AkshayDubey29/MoniFlux/internal/api/handlers"
)

func SetupRouter() *mux.Router {
    router := mux.NewRouter().StrictSlash(true)

    // Define API routes
    router.HandleFunc("/start-test", handlers.StartTest).Methods("POST")
    router.HandleFunc("/schedule-test", handlers.ScheduleTest).Methods("POST")
    router.HandleFunc("/cancel-test", handlers.CancelTest).Methods("POST")
    router.HandleFunc("/restart-test", handlers.RestartTest).Methods("POST")
    router.HandleFunc("/save-results", handlers.SaveResults).Methods("POST")
    router.HandleFunc("/get-all-tests", handlers.GetAllTests).Methods("GET")

    // Middleware can be added here

    return router
}
