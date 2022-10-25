package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

// The struct we are building the json into
type requestPayloadStruct struct {
	ProxyCondition string `json:"proxy_condition"`
}

// Fetch  available addresses from the env file
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// The final address we are fowarding the request to
func getListenAddress() string {
	port := getEnv("PORT", "1338")
	return ":" + port
}

// The "menu" options
func logSetup() {
	a_condtion_url := os.Getenv("A_CONDITION_URL")
	b_condtion_url := os.Getenv("B_CONDITION_URL")
	default_condtion_url := os.Getenv("DEFAULT_CONDITION_URL")

	log.Printf("Server will run on: %s\n", getListenAddress())
	log.Printf("Redirecting to A url: %s\n", a_condtion_url)
	log.Printf("Redirecting to B url: %s\n", b_condtion_url)
	log.Printf("Redirecting to Default url: %s\n", default_condtion_url)
}

func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	requestPayload := parseRequestBody(r)

	url := getProxyUrl(requestPayload.ProxyCondition)
	logRequestPayload(requestPayload, url)
	serveReverseProxy(url, w, r)
}

func logRequestPayload(requestionPayload requestPayloadStruct, proxyUrl string) {
	log.Printf("proxy_condition: %s, proxy_url: %s\n", requestionPayload.ProxyCondition, proxyUrl)
}

// switch to know where the proxy condition is pointing to| Returns a string of the proxy condition from the .env file
func getProxyUrl(proxyConditionRaw string) string {
	proxyCond := strings.ToUpper(proxyConditionRaw)

	a_condtion_url := os.Getenv("A_CONDITION_URL")
	b_condtion_url := os.Getenv("B_CONDITION_URL")
	default_condtion_url := os.Getenv("DEFAULT_CONDITION_URL")

	switch proxyCond {
	case "A":
		return a_condtion_url

	case "B":
		return b_condtion_url

	default:
		return default_condtion_url
	}
}
func main() {
	logSetup()
	//route to execute the handler
	http.HandleFunc("/", handleRequestAndRedirect)
	err := http.ListenAndServe(getListenAddress(), nil)
	if err != nil {
		panic(err)
	}
}

// read the request body and decode | returns a json decoder
func requestBodyDecoder(r *http.Request) *json.Decoder {
	b, err := io.ReadAll(r.Body)
	if err != nil {

		panic(err)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(b))
	return json.NewDecoder(io.NopCloser(bytes.NewBuffer(b)))
}

func parseRequestBody(r *http.Request) requestPayloadStruct {
	decoder := requestBodyDecoder(r)
	var requestPayload requestPayloadStruct
	// decode the json into the requestPayloadStruct
	err := decoder.Decode(&requestPayload)
	if err != nil {
		panic(err)
	}
	// return the ready to use requestPayloadStruct
	return requestPayload

}

// Serve a reverse proxy for a given url-target
func serveReverseProxy(target string, w http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse(target)

	//create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)
	//prepare headers we want to send with the fowarded request
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host
	// serve
	proxy.ServeHTTP(w, r)
}
