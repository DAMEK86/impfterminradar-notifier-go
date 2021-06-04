package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func StartHealthEndpoint() {
	r := mux.NewRouter()
	r.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	fmt.Println("start router on port 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println(err.Error())
	}
}
