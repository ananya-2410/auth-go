package main

import (
	"context"
	"fmt"
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
	user, err := authenticator.Authenticate(r)
	if err != nil {
		log.Fatalln(err)
	}
	hasuraVariables := map[string]string{
		"X-Hasura-Role":    user.UserName(), // result.role
		"X-Hasura-User-Id": user.ID(),
	}

	body := fmt.Sprintf("Hasura Variables: %s \n", hasuraVariables)

	w.Write([]byte(body))
}
func get_user(username string) *auth.DefaultUser {
	fmt.Println(username)
	if username == "user1" {
		return auth.NewDefaultUser("user1", "1", nil, nil)
	} else if username == "user2" {
		return auth.NewDefaultUser("user2", "2", nil, nil)
	} else {
		return auth.NewDefaultUser("anonymous", "0", nil, nil)
	}
}
func createToken(w http.ResponseWriter, r *http.Request) {
	token := uuid.New().String()
	vars := mux.Vars(r)
	user_name := vars["username"]
	// user_id := r.URL.Query().Get("id")
	// user_name := r.URL.Query().Get("username")
	user := get_user(user_name)

	tokenStrategy := authenticator.Strategy(bearer.CachedStrategyKey)
	auth.Append(tokenStrategy, token, user, r)
	body := fmt.Sprintf("token: %s \n", token)
	w.Write([]byte(body))
}

func main() {
	setupGoGuardian()
	router := mux.NewRouter()
	router.HandleFunc("/v1/login/{username}", middleware(http.HandlerFunc(createToken))).Methods("GET")
	router.HandleFunc("/v1/verify", middleware(http.HandlerFunc(getHasuraVariables))).Methods("GET")
	log.Println("Server started and listening on http://127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8081", router)
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
	if userName == "user1" && password == "user1" {
		return auth.NewDefaultUser("user1", "1", nil, nil), nil
	}
	if userName == "user2" && password == "user2" {
		return auth.NewDefaultUser("user2", "2", nil, nil), nil
	}

	return nil, fmt.Errorf("Invalid credentials")
}

func middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing Auth Middleware")
		user, err := authenticator.Authenticate(r)
		if err != nil {
			// 	code := http.StatusNotFound
			fmt.Fprint(w, "ERROR 404 \t")
			// 	http.Error(w, http.StatusText(code), code)
			// 	return
		}
		log.Printf("User %s %s Authenticated\n", user.UserName(), user.ID())
		next.ServeHTTP(w, r)
	})
}
