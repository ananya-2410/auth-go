package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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
	"user1": "1",
	"user2": "2",
	"user0": "0",
}

type input_struct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type token_struct struct {
	Headers Hd `json:"headers"`
}

type Hd struct {
	Token string `json:"token"`
}

func createToken(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	var info input_struct
	err = json.Unmarshal(body, &info)
	if err != nil {
		panic(err)
	}
	token := uuid.New().String()
	username := info.Username
	pwd := info.Password
	if creds_map[username] != pwd {
		log.Fatalln("ERROR CREDENTIALS")
	}
	token_map[token] = string(username)
	log.Println(token)

	output_map := map[string]string{
		"token": token,
	}
	token_dict, _ := json.Marshal(output_map)

	output_body := fmt.Sprintf("\n%s\n", string(token_dict))
	w.Write([]byte(output_body))
}

func getHasuraVariables(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))
	token_temp := &token_struct{}

	err = json.Unmarshal([]byte(body), token_temp)

	if err != nil {
		log.Fatalln(err)
	}

	token_val := token_temp.Headers.Token
	log.Println(string(token_val))
	username, ok := token_map[token_val]
	if !ok {
		username = "user0"
	}
	dict := map[string]string{
		"X-Hasura-Role":    username,
		"X-Hasura-User-Id": role_map[username],
	}
	json_object, err := json.Marshal(dict)
	if err != nil {
		log.Fatalln(err)
	}

	token_map_dict, _ := json.Marshal(token_map)
	fmt.Fprintf(w, string(json_object))

	log.Println(string(json_object))
	log.Println(string(token_map_dict))
}

func main() {

	router := mux.NewRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + "8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	router.HandleFunc("/v1/login", createToken).Methods("POST")
	router.HandleFunc("/v1/verify", getHasuraVariables).Methods("POST")
	log.Println("Server started and listening on http://127.0.0.1:8000")
	log.Fatal(srv.ListenAndServe())
}
