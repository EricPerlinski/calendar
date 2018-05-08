

package main

import (
	"os"
	"fmt"
	"html/template"
	"net/http"
	"log"
)

var (
	HTTP_PORT  = ":8080"
	HTTPS_PORT = ":8443"
	SSL_DIR	   = os.Getenv("CALENDAR_SSL")
	PRIV_KEY   = SSL_DIR+"/privkey.pem"
	PUBLIC_KEY = SSL_DIR+"/cert.pem"
)


func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Somebody connected")
	http.Redirect(w,r, "/index/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Somebody is in index !")
	templateFile := template.Must(template.ParseFiles("templates/index.tmpl"))
	
	
	err := templateFile.Execute(w, nil)
	if err != nil {
		fmt.Println("Error while executing template ! " + err.Error())
		fmt.Fprint(w, "Error while executing template !")
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	
	fmt.Println(r.Method)
	templateFile := template.Must(template.ParseFiles("templates/login.tmpl"))
	
	err := templateFile.Execute(w, nil)
	if err != nil {
		fmt.Println("Error while executing template ! " + err.Error())
		fmt.Fprint(w, "Error while executing template ! " + err.Error())
	}
	
	switch r.Method {

	case "GET" :  
		fmt.Println("About to login !")
	case "POST" :	
		fmt.Println("Checking logging !")
		if r.Method == "POST" {
			r.ParseForm()
			fmt.Println("username:" , r.Form["username"])
			fmt.Println("password:" , r.Form["password"])
		}
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
}

func Run() chan error {

	errs := make(chan error)

	// Starting HTTP server
	go func() {
		log.Printf("Staring HTTP service on %s ...", HTTP_PORT)

		if err := http.ListenAndServe(HTTP_PORT, nil); err != nil {
			errs <- err
		}

	}()

	// Starting HTTPS server
	go func() {
		log.Printf("Staring HTTPS service on %s ...", HTTPS_PORT)
		if err := http.ListenAndServeTLS(HTTPS_PORT, PUBLIC_KEY, PRIV_KEY, nil); err != nil {
			errs <- err
		}
	}()

	return errs
}


func main() {
	fmt.Println("Starting calendar server ... ")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/index/", indexHandler)
	http.HandleFunc("/login/", loginHandler)
	
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Launch server
	errs := Run()

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}

}
