package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

func main() {
	// Connexion à la base de données MySQL
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/forum")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Routes
	http.HandleFunc("/", homePage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/signup/create", signup)
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))

	// Serveur HTTP
	fmt.Println("Serveur démarré sur le port :8080")
	http.ListenAndServe(":8080", nil)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Page d'accueil")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Page de connexion")
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
	_, err = db.Exec("INSERT INTO users (nom, prenom, email, mot_de_passe, pseudo) VALUES (?, ?, ?, ?, ?)", nom, prenom, email, motDePasse, pseudo)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	fmt.Fprintf(w, "Utilisateur créé avec succès !")
}
