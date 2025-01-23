package common

import (
	"encoding/json"
	"net/http"


)

//Marshalling response into JSON 
// Transforming the data into a specific format that can be efficiently stored, transmitted, or processed. 
func WriteJSON(w http.ResponseWriter, status int, data any) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  json.NewEncoder(w).Encode(data)
}

//Marshalling response into JSON 
// converting encoded data back into its original form or a readable format.
func READJSON(r *http.Request, data any) error {
	return json.NewDecoder(r.Body).Decode(data)
}

// function will write a json and return error message
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
