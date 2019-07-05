package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var port string
var endpoint string
var poolID string
var clientID string
var region string
var authClient *CognitoAppClient

func handleRequest(w http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse(endpoint)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = &transport{http.DefaultTransport}

	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host

	proxy.ServeHTTP(w, r)
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
	var mware = []middleware{
		TimerMiddleware,
		ValidateRequestMiddleware,
		AuthenticationMiddleware,
		RequestLoggerMiddleware,
		RequestTransformerMiddleware,
		// CompressionMiddleware,
	}

	http.Handle("/", buildChain(http.HandlerFunc(handleRequest), mware...))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
