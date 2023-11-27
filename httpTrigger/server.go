package httptrigger

import (
	"fmt"
	"net/http"
)

func RunServer() {
	http.HandleFunc("/run-name", HandleNameEvent)

	fmt.Println("Listening on http://localhost:6000")
	if err := http.ListenAndServe(":6000", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
