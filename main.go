package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var port string
var endpoint string
var poolID string
var clientID string
var region string
var authClient *CognitoAppClient

func proxyErrorResponse(status int, message string, res http.ResponseWriter, start time.Time) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(map[string]interface{}{"error": message, "data": nil})
	log.Printf("elasped time %s", time.Since(start))
}

func handleRequest(res http.ResponseWriter, req *http.Request) {
	start := time.Now()
	if _, err := logIncoming(req); err != nil {
		log.Println(err)
		proxyErrorResponse(http.StatusInternalServerError, "Internal server error", res, start)
		return
	}

	token := req.Header.Get("Authorization")

	if token == "" {
		log.Println("no authorization header present in request")
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	user, err := authClient.authenticate(token)
	if err != nil {
		log.Println(err)
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	if user.isAuthenticated == false {
		log.Println("invalid token")
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	formattedUserAttributes, err := json.Marshal(user.cognitoUserAttributes)
	if err != nil {
		log.Println(err)
		proxyErrorResponse(http.StatusInternalServerError, "Internal server error", res, start)
		return
	}

	req.Header.Set("Authorization", string(formattedUserAttributes))
	serveReverseProxy(endpoint, res, req, start)
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request, start time.Time) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	log.Printf("elasped time %s", time.Since(start))
	proxy.ServeHTTP(res, req)
}

func init() {
	port = getEnv("PORT")
	endpoint = getEnv("URL")
	poolID = getEnv("POOL_ID")
	clientID = getEnv("CLIENT_ID")
	region = getEnv("REGION")

	config := &CognitoAppClientConfig{
		Region:   region,
		PoolID:   poolID,
		ClientID: clientID,
	}

	log.Println("initializing cognito client")
	client, err := NewCognitoAppClient(config)

	if err != nil {
		panic(err)
	}

	authClient = client
}

func main() {
	http.HandleFunc("/", handleRequest)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
