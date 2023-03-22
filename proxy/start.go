package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Start() {

	log.Fatal(http.ListenAndServe(":80", nil))
}

func RegisterHttpHandler(jobs <-chan int) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Modify the request before sending it to the backend server
		r.URL.Path = "/newpath" // replace the path with a new path
		backendUrl, err := url.Parse("http://localhost:8081")
		if err != nil {
			log.Fatal(err)
		}
		r.Host = backendUrl.Host // set the backend host as the request host

		// Proxy the modified request to the backend server
		reverseProxy := httputil.NewSingleHostReverseProxy(backendUrl)
		reverseProxy.ServeHTTP(w, r)
	})

}
