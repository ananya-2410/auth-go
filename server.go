package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	client := http.Client{}
	// Pass username password to auth0 to get token
	req, err := http.NewRequest("GET", "/webhook/login", nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("username", "abcd")
	req.Header.Set("password", "pwd12345")

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	// Get token
	token := string(body)
	// Use above token to verify
	req1, err := http.NewRequest("GET", "/webhook/verify", nil)
	if err != nil {
		log.Fatalln(err)
	}
	req1.Header.Set("Authorization", token)
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := client.Do(req1)

	if err != nil {
		log.Fatalln(err)
		log.Printf("X-Hasura-Role: Anonymous")
	}
	defer resp1.Body.Close()
	// Get Role & User ID if token is valid
	role := resp1.Header.Get("X-Hasura-Role")
	user_id := resp1.Header.Get("X-Hasura-User-Id")

	// Print role and user id
	log.Printf(role)
	log.Printf(user_id)

}
