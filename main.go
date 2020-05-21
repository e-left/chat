package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/stretchr/objx"

	"github.com/e-left/chat/trace"
	"github.com/gorilla/mux"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

//templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "The address of the application")
	flag.Parse()

	// Oauth2 setup
	gomniauth.SetSecurityKey("TESTING")
	gomniauth.WithProviders(
		google.New("795839929081-sspqj0je1kbnac77e25lleuic56nddbs.apps.googleusercontent.com", "    5OENW-iLFshrZw5aHiCjghqD", "http://localhost:8080/auth/callback/google"),
		github.New("4f35d044c2c972f59a8d", "e59be4455baa54900f35e9d79437229f5fc76806", "http://localhost:8080/auth/callback/github"),
	)

	r := newRoom()
	// this line enables tracing. if deleted program prints nothing
	r.tracer = trace.New(os.Stdout)

	// mux handling
	router := mux.NewRouter()
	//router.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	// new static file handler for gorilla mux
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	router.Handle("/login", &templateHandler{filename: "login.html"})
	router.HandleFunc("/auth/{action}/{service}", loginHandler)
	router.Handle("/room", r)
	go r.run()

	log.Println("Starting webserver on", *addr)
	if err := http.ListenAndServe(*addr, router); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
