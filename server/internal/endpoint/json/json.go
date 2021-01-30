package json

import (
	"encoding/json"
	"net/http"

	"github.com/fhodun/BitChat/server/internal/util"
)

const searchPath = "/messages/search/"
const userPath = "/messages/user/"
const allPatch = "/messages/all"

// Start dupa
func Start() {
	properties := util.LoadConfig()

	http.HandleFunc(searchPath, searchMessages)
	http.HandleFunc(userPath, userMessages)
	http.HandleFunc(allPatch, allMessages)

	err := http.ListenAndServe(":"+properties.JSONEndpointPort, nil)
	util.CheckForError(err, "Can't create JSON endpoint")
}

func searchMessages(w http.ResponseWriter, r *http.Request) {
	var searchTerm = r.URL.Path[len(searchPath):]

	returnQuery("message", searchTerm, "", w, r)
}

func userMessages(w http.ResponseWriter, r *http.Request) {
	var username = r.URL.Path[len(userPath):]

	returnQuery("message", "", username, w, r)
}

func allMessages(w http.ResponseWriter, r *http.Request) {
	returnQuery("message", "", "", w, r)
}

func returnQuery(actionType string, search string, username string,
	w http.ResponseWriter, r *http.Request) {

	actions := util.QueryMessages(actionType, search, username)
	payload, err := json.Marshal(actions)
	util.CheckForError(err, "Can't create JSON response")

	w.Header().Set("Content-Type", "text/json")
	w.Write(payload)
}
