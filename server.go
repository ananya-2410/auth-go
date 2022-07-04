package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type input_struct struct {
	username string
	password string
}

type token_struct struct {
	token string
}

func createToken(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	var input input_struct
	err = json.Unmarshal(body, &input)
	if err != nil {
		panic(err)
	}
	token := uuid.New().String()
	username := input.username
	pwd := input.password
	if creds_map[username] != pwd {
		fmt.Println("Error credentials")
		log.Fatalln("ERROR CREDENTIALS")
	}
	token_map[token] = username
	log.Println(token)

	output_map := map[string]string{
		"token": token,
	}
	token_dict, _ := json.Marshal(output_map)
	output_body := fmt.Sprintf("%s", string(token_dict))
	w.Write([]byte(output_body))
}

func getHasuraVariables(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	var token_temp token_struct
	err = json.Unmarshal(body, &token_temp)
	if err != nil {
		panic(err)
	}
	token_val := token_temp.token
	username, ok := token_map[token_val]
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
	// token_map_dict, _ := json.Marshal(token_map)
	output_body := fmt.Sprintf("\n %s", string(json_object))
	w.Write([]byte(output_body))
	log.Println(string(json_object))
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/v1/login", createToken).Methods("POST")
	router.HandleFunc("/v1/verify", getHasuraVariables).Methods("POST")
	log.Println("Server started and listening on http://127.0.0.1:8000")
	http.ListenAndServe("127.0.0.1:8081", router)
}
