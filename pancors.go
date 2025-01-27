package pancors

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type corsTransport struct {
	referer     string
	origin      string
	credentials string
}

func (t corsTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Put in the Referer if specified
	if t.referer != "" {
		r.Header.Add("Referer", t.referer)
	}

	// Do the actual request
	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	res.Header.Set("Access-Control-Allow-Origin", t.origin)
	res.Header.Set("Access-Control-Allow-Credentials", t.credentials)

	return res, nil
}

func handleProxy(w http.ResponseWriter, r *http.Request, origin string, credentials string) {
	// Handle with PreFlight
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}
	// Check for the User-Agent header
	if r.Header.Get("User-Agent") == "" {
		http.Error(w, "Missing User-Agent header", http.StatusBadRequest)
		return
	}

	// Get the optional Referer header
	referer := r.URL.Query().Get("referer")
	if referer == "" {
		referer = r.Header.Get("Referer")
	}
	
	// Get the target endpoint in Headers
	urlHost := r.Header.Get("target")
	urlPath := r.URL.Path // pass the path to target endpoint
	urlParam := urlHost + urlPath
	if urlHost == "" || urlPath == "" {
		// Get the URL in params
		urlParam = r.URL.Query().Get("url")
	}

	// Validate the URL
	urlParsed, err := url.Parse(urlParam)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	// Check if HTTP(S)
	if urlParsed.Scheme != "http" && urlParsed.Scheme != "https" {
		http.Error(w, "The URL scheme is neither HTTP nor HTTPS", http.StatusBadRequest)
		return
	}

	// Setup for the proxy
	proxy := httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL = urlParsed
			r.Host = urlParsed.Host
		},
		Transport: corsTransport{referer, origin, credentials},
	}

	// Execute the request
	proxy.ServeHTTP(w, r)
}

// HandleProxy is a handler which passes requests to the host and returns their
// responses with CORS headers
func HandleProxy(w http.ResponseWriter, r *http.Request) {
	handleProxy(w, r, "*", "true")
}

// HandleProxyFromHosts is a handler which passes requests only from specified to the host
func HandleProxyWith(origin string, credentials string) func(http.ResponseWriter, *http.Request) {
	if !(credentials == "true" || credentials == "false") {
		log.Panicln("Access-Control-Allow-Credentials can only be 'true' or 'false'")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		handleProxy(w, r, origin, credentials)
	}
}
