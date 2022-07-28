package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

func (s *server) LandingPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	s.templates.LandingPage.Execute(w, nil)
}

func (s *server) GetBoard(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	var content string
	var signature string
	var timestamp string
	if !IsValidKey(key) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if key == "ab589f4dde9fce4180fcf42c7b05185b0a02a5d682e353fa39177995083e0583" {
		content, signature, timestamp = GenerateFakePage()
	} else {
		if OnDenyList(key) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "sorry, this key is on the deny list")
			return
		}
		if !HasBeenModified(s, r, key) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		res, err := s.kv.Get(r.Context(), key+":content").Result()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		content = res
		res, err = s.kv.Get(r.Context(), key+":signature").Result()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		signature = res
		res, err = s.kv.Get(r.Context(), key+":timestamp").Result()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		timestamp = res
	}
	t, _ := time.Parse(time.RFC3339, timestamp)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Spring-Signature", signature)
	w.Header().Set("ETag", signature)
	w.Header().Set("Spring-Version", "83")
	w.Header().Set("Last-Modified", t.Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
	if r.Header.Get("Spring-Version") == "" {
		s.templates.BoardPage.Execute(w, struct {
			Content template.HTML
			Key     string
		}{
			Content: template.HTML(content),
			Key:     key,
		})
		return
	}
	w.Write([]byte(content))
}

func HasBeenModified(s *server, r *http.Request, key string) bool {
	if r.Header.Get("If-Modified-Since") == "" {
		return true
	}
	res := s.kv.Get(r.Context(), key+":timestamp")
	if res.Err() != nil {
		return false
	}

	t, _ := time.Parse(time.RFC3339, res.String())
	if t.Format(http.TimeFormat) == r.Header.Get("If-Modified-Since") {
		return false
	}
	return true
}

func (s *server) ChangeBoardContent(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if !IsValidKey(key) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "this is an invalid key")
		return
	}
	if OnDenyList(key) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "sorry, this key is on the deny list")
		return
	}

	sh, err := SpringHeaders(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if len(body) > 2217 {
		w.WriteHeader(413)
		fmt.Fprintln(w, "content out of spec. max size is 2217")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error reading body")
		return
	}

	if !ValidateKeyAndSignature(key, sh.Signature, string(body)) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "invalid signature")
		fmt.Println(body)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error parsing body")
		return
	}
	stamp, exists := doc.Find("time").Attr("datetime")
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "no time stamp")
		return
	}
	if v, t, err := ValidTimestamp(stamp); !v || err != nil {
		if t.Before(time.Now()) {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintln(w, "timestamp is in the past")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		if err != nil {
			fmt.Fprintf(w, "encountered err: %s\n", err.Error())
		}
		fmt.Fprintln(w, "invalid time stamp")
		return
	}

	res := s.kv.Set(r.Context(), key+":content", string(body), time.Hour*24*22)
	if res.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error saving content")
		return
	}
	res = s.kv.Set(r.Context(), key+":signature", sh.Signature, time.Hour*24*22)
	if res.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error saving content")
		return
	}
	res = s.kv.Set(r.Context(), key+":timestamp", stamp, time.Hour*24*22)
	if res.Err() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error saving content")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "success")

}
