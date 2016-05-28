package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

var portmap = make(map[string]string)

func init() {
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
	for i, l := range ss {
		if l == "" {
			continue
		}
		ll := strings.Split(l, " ")
		if len(ll) != 2 {
			fmt.Printf("PORTMAP: not a valid line: %v\n%v\n", i, l)
			os.Exit(1)
		}
		hostprefix, port := ll[0], ll[1]
		portmap[hostprefix] = port
	}
	if len(portmap) == 0 {
		fmt.Println("not any redirection specfied. nothing to do.")
		os.Exit(1)
	}
	_, ok := portmap["_"]
	if !ok {
		fmt.Println("_ will bind all prefix except specified. should exist")
		os.Exit(1)
	}
	fmt.Println(portmap)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		director := func(req *http.Request) {
			port, _ := portmap["_"]
			hh := strings.Split(r.Host, ".")
			if len(hh) == 3 {
				p, ok := portmap[hh[0]]
				if ok {
					port = p
				}
			}
			req.URL.Scheme = "http"
			req.URL.Host = "localhost:" + port
			req.URL.Path = r.URL.Path
			req.URL.RawQuery = r.URL.RawQuery
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(":80", nil))
}
