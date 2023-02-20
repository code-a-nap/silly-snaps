package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.Redirect(w, r, "/", 301)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")
		if email == "user@silly-snaps.com" && password == os.Getenv("USER_PWD") {
			token, _ := GenerateJWT(email, "user", "flag{your_secret_flag}")
			cookie := &http.Cookie{
				Name:   "token",
				Value:  token,
				MaxAge: 3600,
				Secure: false,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", 301)
		} else if email == "admin@silly-snaps.com" && password == os.Getenv("ADMIN_PWD") {
			token, _ := GenerateJWT(email, "admin", os.Getenv("FLAG"))
			cookie := &http.Cookie{
				Name:   "token",
				Value:  token,
				MaxAge: 3600,
				Secure: false,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", 301)
		} else {
			w.Write([]byte("Sorry, Unauthorized..."))
		}
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
		return
	}

	claims, err2 := ValidateToken(tokenCookie.Value)
	if err2 != nil {
		w.Write([]byte("Sorry, Your token is not valid.."))
		return
	}

	tmpl := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).ParseFiles("home.html"))
	tmpl.Execute(w, *claims)

}

func addPictureHandler(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
		return
	}

	_, err2 := ValidateToken(tokenCookie.Value)
	if err2 != nil {
		w.Write([]byte("Sorry, Your token is not valid.."))
		return
	}

	// Feature still under development. For now just redirect.
	http.Redirect(w, r, r.URL.Query().Get("redirect"), 301)
}

func reportToAdmin(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, nil)
		return
	}

	_, err2 := ValidateToken(tokenCookie.Value)
	if err2 != nil {
		w.Write([]byte("Sorry, Your token is not valid.."))
		return
	}

	url := r.URL.Query().Get("url")
	_, err3 := exec.Command("node", "headless.js", url).Output()
	if err3 != nil {
		log.Printf("Error %s while executing command with url: %s\n", err3, url)
	}

	tmpl := template.Must(template.New("report.html").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).ParseFiles("report.html"))
	tmpl.Execute(w, nil)

}

func main() {

	err := checkEnvVars()
	if err != nil {
		panic(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/addPicture", addPictureHandler)
	router.HandleFunc("/report", reportToAdmin)

	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	router.Use(loggingMiddleware)

	openLogFile("logs/access.log")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	http.ListenAndServe(":"+port, router)
}
