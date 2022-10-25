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

type requestPayloadStruct struct {
	ProxyCondition string `json:"proxy_condition"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getListenAddress() string {
	port := getEnv("PORT", "1338")
	return ":" + port
}

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
	http.HandleFunc("/", handleRequestAndRedirect)
	err := http.ListenAndServe(getListenAddress(), nil)
	if err != nil {
		panic(err)
	}
}

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
	err := decoder.Decode(&requestPayload)
	if err != nil {
		panic(err)
	}
	return requestPayload

}

func serveReverseProxy(target string, w http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = url.Host
	proxy.ServeHTTP(w, r)
}
