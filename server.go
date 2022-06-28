package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shaj13/go-guardian/auth"
	"github.com/shaj13/go-guardian/auth/strategies/basic"
	"github.com/shaj13/go-guardian/auth/strategies/bearer"
	"github.com/shaj13/go-guardian/store"
)

var authenticator auth.Authenticator
var cache store.Cache

func getHasuraVariables(w http.ResponseWriter, r *http.Request) {
	// hasuraVariables := map[string]string{
	// 	"X-Hasura-Role":    "user", // result.role
	// 	"X-Hasura-User-Id": "1",
	// }

	// body := fmt.Sprintf("Hasura Variables: %s \n", hasuraVariables)
	r.Header.Set("X-Hasura-Role", "user")
	r.Header.Set("X-Hasura-User-Id", "1")
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(r)

	if err != nil {

		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(string(body))
		log.Fatalln(err)
	}
	output_body := string(body)
	log.Printf(output_body)
	w.Write([]byte(body))
}

func createToken(w http.ResponseWriter, r *http.Request) {
	token := uuid.New().String()
	user := auth.NewDefaultUser("medium", "1", nil, nil)
	tokenStrategy := authenticator.Strategy(bearer.CachedStrategyKey)
	auth.Append(tokenStrategy, token, user, r)
	body := fmt.Sprintf("token: %s \n", token)
	w.Write([]byte(body))
}

func main() {
	setupGoGuardian()
	router := mux.NewRouter()
	router.HandleFunc("/v1/login", middleware(http.HandlerFunc(createToken))).Methods("GET")
	router.HandleFunc("/v1/verify", middleware(http.HandlerFunc(getHasuraVariables))).Methods("GET")
	log.Println("Server started and listening on http://127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8080", router)
}

func setupGoGuardian() {
	authenticator = auth.New()
	cache = store.NewFIFO(context.Background(), time.Minute*10)

	basicStrategy := basic.New(validateUser, cache)
	tokenStrategy := bearer.New(bearer.NoOpAuthenticate, cache)

	authenticator.EnableStrategy(basic.StrategyKey, basicStrategy)
	authenticator.EnableStrategy(bearer.CachedStrategyKey, tokenStrategy)
}

func validateUser(ctx context.Context, r *http.Request, userName, password string) (auth.Info, error) {
	if userName == "medium" && password == "medium" {
		return auth.NewDefaultUser("medium", "1", nil, nil), nil
	}

	return nil, fmt.Errorf("Invalid credentials")
}

func middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing Auth Middleware")
		user, err := authenticator.Authenticate(r)
		if err != nil {
			code := http.StatusNotFound
			fmt.Fprint(w, "ERROR 404 \t")
			http.Error(w, http.StatusText(code), code)
			return
		}
		log.Printf("User %s Authenticated\n", user.UserName())
		next.ServeHTTP(w, r)
	})
}
