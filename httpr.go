package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	portmap := make(map[string]string)
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
}

func main() {
	http.HandleFunc("/", redirect)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	log.Print(r.URL.Path)
	port, _ := portmap["_"]
	hh := strings.Split(r.URL.Host, ".")
	if len(hh) == 3 {
		p, ok := portmap[hh[0]]
		if ok {
			port = p
		}
	}
	req, err := http.NewRequest(r.Method, "http://localhost:"+port+r.URL.Path, r.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("proxy: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("proxy: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
