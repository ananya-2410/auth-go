package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	token_map = make(map[string]string)
)

var creds_map = map[string]string{
	"user1": "user1",
	"user2": "user2",
}

var role_map = map[string]string{
	"user1":     "1",
	"user2":     "2",
	"anonymous": "0",
}

type Foo struct {
	Name string `json:"X-Hasura-Role"`
	Id   string `json:"X-Hasura-User-Id"`
}

func getHasuraVariables(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Please pass the data as URL form encoded", http.StatusBadRequest)
		return
	}
	token := r.Form.Get("token")
	username, ok := token_map[token]
	if !ok {
		username = "anonymous"
	}
	dict := map[string]string{
		"X-Hasura-Role":    username,
		"X-Hasura-User-Id": role_map[username],
	}
	json_object, err := json.Marshal(dict)
	if err != nil {
		log.Fatalln(err)
	}
	body := fmt.Sprintf("\n %s\n ", string(json_object))
	w.Write([]byte(body))
	log.Println(string(json_object))
}
func main() {
	router := mux.NewRouter()

	router.HandleFunc("/v1/login", createToken).Methods("POST")
	router.HandleFunc("/v1/verify", getHasuraVariables).Methods("POST")
	log.Println("Server started and listening on http://127.0.0.1:8000")
	http.ListenAndServe("127.0.0.1:8081", router)
}

func createToken(w http.ResponseWriter, r *http.Request) {

	token := uuid.New().String()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Please pass the data as URL form encoded", http.StatusBadRequest)
		return
	}
	username := r.Form.Get("username")
	pwd := r.Form.Get("password")
	if creds_map[username] != pwd {
		fmt.Println("Error credentials")
		log.Fatalln("ERROR CREDENTIALS")
	}
	body := fmt.Sprintf("token: %s \n", token)
	token_map[token] = username
	w.Write([]byte(body))
}
