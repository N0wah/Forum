package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var db *sql.DB
var err error

var store = sessions.NewCookieStore([]byte("secret"))

func main() {
	// Connexion à la base de données MySQL
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/forum")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Routes
	http.HandleFunc("/", isAuthenticated(homePage))
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/authenticate", authenticate)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/signup/create", signup)
	http.HandleFunc("/signup/success", signupSuccess)
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))

	// Serveur HTTP
	fmt.Println("Serveur démarré sur le port :8080")
	http.ListenAndServe(":8080", nil)
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérification de la session
		session, err := store.Get(r, "session")
		if err != nil || session.Values["pseudo"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	// Suppression de la session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["pseudo"] = nil
	session.Options.MaxAge = -1 // Supprime la session
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirection vers la page de connexion
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}


func homePage(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFiles("pages/index.html"))
	tpl.Execute(w, nil)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFiles("pages/login.html"))
	tpl.Execute(w, nil)
}

func signupPage(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFiles("pages/register.html"))
	tpl.Execute(w, nil)
}

func signup(w http.ResponseWriter, r *http.Request) {
	// Récupération des données du formulaire
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nom := r.Form.Get("nom")
	prenom := r.Form.Get("prenom")
	email := r.Form.Get("email")
	motDePasse := r.Form.Get("motdepasse")
	pseudo := r.Form.Get("pseudo")

	// Insertion des données dans la base de données
	_, err = db.Exec("INSERT INTO utilisateurs (nom, prenom, email, mot_de_passe, pseudo) VALUES (?, ?, ?, ?, ?)", nom, prenom, email, motDePasse, pseudo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/signup/success", http.StatusSeeOther)
}

func signupSuccess(w http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseFiles("pages/confirmation_register.html"))
	tpl.Execute(w, nil)
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	email := r.Form.Get("email")
	motDePasse := r.Form.Get("motdepasse")

	// Vérification de l'utilisateur dans la base de données
	var pseudo string
	err = db.QueryRow("SELECT pseudo FROM utilisateurs WHERE email = ? AND mot_de_passe = ?", email, motDePasse).Scan(&pseudo)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Création de la session
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["pseudo"] = pseudo
	session.Save(r, w)

	// Redirection vers la page d'accueil ou une autre page sécurisée
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
