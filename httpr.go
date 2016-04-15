package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	b, err := ioutil.ReadFile("PORTMAP")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Print("please make PORTMAP file.")
			os.Exit(1)
		}
		log.Fatal(err)
	}
	s := strings.TrimRight(string(b), "\r\n")
	ss := strings.Split(s, "\n")
	portmap := make(map[string]string)
	for i, l := range ss {
		if l == "" {
			continue
		}
		ll := strings.Split(l, ":")
		if len(ll) != 2 {
			log.Fatalf("PORTMAP: not a valid line: %v\n%v", i, l)
		}
		portmap[ll[0]] = ll[1]
	}
	fmt.Print(portmap)
	if len(portmap) == 0 {
		fmt.Print("not any redirection specfied. nothing to do.")
	}
	for prefix, port := range portmap {
		http.HandleFunc(prefix, redirect(prefix, port))
	}
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func redirect(prefix, port string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest(r.Method, "http://localhost:"+port+"/"+strings.TrimPrefix(r.URL.Path, prefix), r.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		t, err := template.New("resp").Parse("{{.}}")
		if err != nil {
			log.Fatal(err)
		}
		t.Execute(w, template.HTML(string(body)))
	}
}
