package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Take one handler function and wrap it with another
type middleware func(http.HandlerFunc) http.HandlerFunc

// Build the middleware chain recursively
func buildChain(f http.HandlerFunc, m ...middleware) http.HandlerFunc {
	// Last middleware in chain is returned
	if len(m) == 0 {
		return f
	}

	// Otherwise, nest the functions
	return m[0](buildChain(f, m[1:cap(m)]...))
}

// TimerMiddleware - log how long the execution takes
var TimerMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		log.Printf("starting a new request: %s", requestID)
		start := time.Now()
		f(w, r)
		end := time.Now()
		log.Printf("finished with request: %s", requestID)
		log.Printf("total execution time: %s", end.Sub(start))
	}
}

// ValidateRequestMiddleware - checks that request body is valid JSON
var ValidateRequestMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				log.Println(err)
				log.Println("unable to read request body for JSON validation")
				errorResponse(http.StatusInternalServerError, "Internal server error", w)
				return
			}

			// Close the body since we read it, and in case it's not valid JSON
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			if !IsJSON(body) {
				log.Println("invalid JSON in request body")
				errorResponse(http.StatusBadRequest, "Body must be valid JSON", w)
				return
			}

			// Set the content type header since we know it is JSON
			r.Header.Set("Content-Type", "application/json")
		}

		f(w, r)
	}
}

// AuthenticationMiddleware - check that the authentication token was provided
var AuthenticationMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			log.Printf("client: %s had no authorization header", r.RemoteAddr)
			errorResponse(http.StatusUnauthorized, "Unauthorized", w)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 {
			log.Printf("client: %s had a malformed token in header", r.RemoteAddr)
			errorResponse(http.StatusUnauthorized, "Unauthorized", w)
			return
		}

		user, err := authClient.authenticate(bearerToken[1])
		if err != nil {
			log.Println(err)
			log.Printf("client: %s unexpected authentication error", r.RemoteAddr)
			errorResponse(http.StatusUnauthorized, "Unauthorized", w)
			return
		}

		if user.authenticated == false {
			log.Printf("client: %s has invalid token in header", r.RemoteAddr)
			errorResponse(http.StatusUnauthorized, "Unauthorized", w)
			return
		}

		formattedUserAttributes, err := json.Marshal(user.attributes)
		if err != nil {
			log.Println(err)
			errorResponse(http.StatusInternalServerError, "Internal server error", w)
			return
		}

		r.Header.Set("Authorization", string(formattedUserAttributes))
		f(w, r)
	}
}

// RequestLoggerMiddleware - log the incoming request as a JSON object
var RequestLoggerMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := requestLogItem{
			Host:          r.Host,
			Address:       r.RemoteAddr,
			Headers:       transformHeaders(r.Header),
			Method:        r.Method,
			RequestURI:    r.RequestURI,
			Proto:         r.Proto,
			UserAgent:     r.Header.Get("User-Agent"),
			ContentLength: r.ContentLength,
			Query:         r.URL.Query(),
		}

		if r.Method == "POST" || r.Method == "PUT" {
			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				log.Println(err)
				log.Println("unexpected error when parsing HTTP body")
				errorResponse(http.StatusInternalServerError, "Internal server error", w)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			var placeholder map[string]interface{}
			err = json.Unmarshal(body, &placeholder)

			if err != nil {
				log.Println(err)
				log.Printf("unable to parse JSON in HTTP request body")
			} else {
				item.RequestBody = placeholder
			}
		}

		stringifyAndLog(item)

		f(w, r)
	}
}

// RequestTransformerMiddleware - check that the authentication token was provided
var RequestTransformerMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			body, err := ioutil.ReadAll(r.Body)

			if err != nil {
				log.Println(err)
				log.Println("unexpected error when parsing HTTP body")
				errorResponse(http.StatusInternalServerError, "Internal server error", w)
				return
			}

			// Camelize JSON object keys
			body = convertJSONKeys(json.RawMessage(body), "camel")

			// Recalculate the content lenght
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))
			r.ContentLength = int64(len(body))

			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		f(w, r)
	}
}

// CompressionMiddleware - use gzip to compress response body
var CompressionMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")

			gz := gzPool.Get().(*gzip.Writer)
			defer gzPool.Put(gz)

			gz.Reset(w)
			defer gz.Close()

			f(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
		} else {
			f(w, r)
		}
	}
}
