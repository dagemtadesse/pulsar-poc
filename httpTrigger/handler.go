package httptrigger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

type EventData struct {
	Person string `json:"name"`
}

type ResponseData struct {
	Message string `json:"msg"`
	Data    string `json:"fn_output"`
}

func HandleNameEvent(w http.ResponseWriter, r *http.Request) {
	var person EventData

	// reading the event data from http request
	eventData, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Error reading event data: %s", err)
		return
	}

	// unmarshalling json data
	err = json.Unmarshal(eventData, &person)
	if err != nil {
		fmt.Fprintf(w, "Error parsing event data: %s", err)
		return
	}

	result := functionExecutor("./app/module/index.js", person.Person)

	res, err := json.Marshal(ResponseData{Message: "Function execution success!", Data: result})
	if err != nil {
		fmt.Fprintf(w, "Error marshalling response data: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func functionExecutor(fnPath string, arg string) string {
	// runs the node file by passing any arguments to the function
	cmd := exec.Command("node", "./app/module/index.js", arg)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Error executing command: %s", err)
	}

	return string(output)
}
