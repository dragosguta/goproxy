package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
	log.Printf("elapsed time: %s", time.Since(start))
}

func handleRequest(res http.ResponseWriter, req *http.Request) {
	start := time.Now()
	valid, err := validJSONRequestBody(req)

	if err != nil {
		proxyErrorResponse(http.StatusInternalServerError, "Internal server error", res, start)
		return
	}

	if !valid {
		proxyErrorResponse(http.StatusBadRequest, "Body must be valid JSON", res, start)
		return
	}

	logRequest(req)

	authorizationHeader := req.Header.Get("Authorization")
	if authorizationHeader == "" {
		log.Println("no authorization header present in request")
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	bearerToken := strings.Split(authorizationHeader, " ")
	if len(bearerToken) != 2 {
		log.Println("malformed authorization header present in request")
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	user, err := authClient.authenticate(bearerToken[1])
	if err != nil {
		log.Println(err)
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	if user.authenticated == false {
		log.Println("invalid token")
		proxyErrorResponse(http.StatusUnauthorized, "Unauthorized", res, start)
		return
	}

	formattedUserAttributes, err := json.Marshal(user.attributes)
	log.Println(string(formattedUserAttributes))
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
	proxy.Transport = &transport{http.DefaultTransport}

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
	log.Printf("elapsed time: %s", time.Since(start))
}

func init() {
	port = getEnv("PORT")
	endpoint = getEnv("URL")
	poolID = getEnv("POOL_ID")
	clientID = getEnv("CLIENT_ID")
	region = getEnv("AWS_REGION")

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
	final := http.HandlerFunc(handleRequest)
	http.Handle("/", Gzip(final))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
