

package main

import (
	"os"
	"fmt"
	"html/template"
	"net/http"
	"log"
	"time"
)

var (
	HTTP_PORT  = ":8080"
	HTTPS_PORT = ":8443"
	SSL_DIR	   = os.Getenv("CALENDAR_SSL")
	PRIV_KEY   = SSL_DIR+"/privkey.pem"
	PUBLIC_KEY = SSL_DIR+"/cert.pem"
)

type loginData struct {
	Connected bool
	Username  string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Somebody connected")
	http.Redirect(w,r, "/index/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r*http.Request) {
	fmt.Println("User disconnected")
	c := &http.Cookie{
		Name : "username",
		Value : "",
		Path : "/",
		MaxAge: -1}
	http.SetCookie(w, c)
	http.Redirect(w,r, "/index/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Somebody is in index !")

	cookie, checkCookie := r.Cookie("username")

	var data loginData

	if checkCookie == nil {
		data=loginData{true, cookie.Value}
	} else {
		data=loginData{false, ""}
	}

	templateFile := template.Must(template.ParseFiles("templates/index.tmpl"))
	err := templateFile.Execute(w,data)
	if err != nil {
		fmt.Println("Error while executing template [index]" + err.Error())
		fmt.Fprint(w, "Error while executing template [index]")
	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	
	fmt.Println(r.Method)

	switch r.Method {

	case "GET" :  
		fmt.Println("About to login !")
		templateFile := template.Must(template.ParseFiles("templates/login.tmpl"))
		err := templateFile.Execute(w,nil)
		if err != nil {
			fmt.Println("Error while executing template [login]" + err.Error())
			fmt.Fprint(w, "Error while executing template [login]")
		}
	case "POST" :	
		fmt.Println("Checking logging !")
		
		r.ParseForm()
		
		okForm := checkAndStoreCredentials(w, r.FormValue("username"), r.FormValue("password"))
		 
		if okForm == true {
			http.Redirect(w, r, "/index/", http.StatusSeeOther)
		} else {
			templateFile := template.Must(template.ParseFiles("templates/login.tmpl"))
			err := templateFile.Execute(w,nil)
			if err != nil {
				fmt.Println("Error while executing template [login]" + err.Error())
				fmt.Fprint(w, "Error while executing template [login]")
			}
		}
	default:
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
}

func checkAndStoreCredentials(w http.ResponseWriter, username string, password string) (okForm bool) {

	fmt.Println("Username["+username+"]")
	fmt.Println("Password["+password+"]")
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name : "username", Value : username, Expires: expiration, Path : "/"}
	http.SetCookie(w,&cookie)
	
	return true
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
	http.HandleFunc("/logout/", logoutHandler)
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
