package main

import "net/http"

func corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
		headers.Set("Access-Control-Allow-Headers", "Content-Type, If-Modified-Since, Spring-Signature, Spring-Version")
		headers.Set("Access-Control-Expose-Headers", "Content-Type, Last-Modified, Spring-Signature, Spring-Version")
		//headers.Set("Content-Security-Policy", "default-src 'none'; style-src 'self' 'unsafe-inline'; font-src 'self'; script-src 'self'; form-action *; connect-src *;")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	}
}
