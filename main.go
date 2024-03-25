package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var db *sql.DB
var err error
var store = sessions.NewCookieStore([]byte("secret"))

type User struct {
	Pseudo   string
	Nom      string
	Prenom   string
	Email    string
	Password string
}

type Topic struct {
	ID            int
	Utilisateur   string
	UtilisateurID int
	Titre         string
	Contenu       string
}

type Commentaire struct {
	ID            int
	UtilisateurID int
	TopicID       int
	Contenu       string
}

func main() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/forum")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	http.HandleFunc("/", isAuthenticated(homePage))
	http.HandleFunc("/profil", profilPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/authenticate", authenticate)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/signup/create", signup)
	http.HandleFunc("/signup/success", signupSuccess)
	http.HandleFunc("/update-profile", updateProfile)
	http.HandleFunc("/topics", viewAllTopicsPage)
	http.HandleFunc("/topic/create", createTopicPage)
	http.HandleFunc("/topic/details", viewTopicDetailsPage)

	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))

	fmt.Println("Serveur démarré sur le port :8080")
	http.ListenAndServe(":8080", nil)
}

func isAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session")
		if err != nil || session.Values["pseudo"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["pseudo"] = nil
	session.Options.MaxAge = -1 
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	var pseudo string
	var id int
	err = db.QueryRow("SELECT pseudo, id FROM utilisateurs WHERE email = ? AND mot_de_passe = ?", email, motDePasse).Scan(&pseudo, &id)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["pseudo"] = pseudo
	session.Values["id"] = id
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}


func profilPage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pseudo, ok := session.Values["pseudo"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := getUserData(pseudo)

	tpl := template.Must(template.ParseFiles("pages/profil.html"))
	tpl.Execute(w, user)
}

func getUserData(pseudo string) User {
	var user User
	err := db.QueryRow("SELECT pseudo, nom, prenom, email, mot_de_passe FROM utilisateurs WHERE pseudo = ?", pseudo).Scan(&user.Pseudo, &user.Nom, &user.Prenom, &user.Email, &user.Password)
	if err != nil {
		fmt.Println("Erreur lors de la récupération des données de l'utilisateur:", err)
	}
	return user
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pseudo, ok := session.Values["pseudo"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newPseudo := r.Form.Get("pseudo")
	newNom := r.Form.Get("nom")
	newPrenom := r.Form.Get("prenom")
	newEmail := r.Form.Get("email")
	newPassword := r.Form.Get("password")

	_, err = db.Exec("UPDATE utilisateurs SET pseudo=?, nom=?, prenom=?, email=?, mot_de_passe=? WHERE pseudo=?", newPseudo, newNom, newPrenom, newEmail, newPassword, pseudo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/logout", http.StatusSeeOther)
}

func getAllTopics() ([]Topic, error) {
	var topics []Topic

	rows, err := db.Query("SELECT id, utilisateurs_pseudo, utilisateur_id, titre, contenu FROM topics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic
		err := rows.Scan(&topic.ID, &topic.Utilisateur, &topic.UtilisateurID, &topic.Titre, &topic.Contenu)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func createTopicPage(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.Values["pseudo"] == nil || session.Values["id"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Convertir l'ID de l'utilisateur en int
	pseudo, ok := session.Values["pseudo"]
	id, ok2 := session.Values["id"].(int)
	if !ok {
		http.Error(w, "ID utilisateur invalide", http.StatusInternalServerError)
		return
	}

	if !ok2 {
		http.Error(w, "ID utilisateur invalide", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		titre := r.Form.Get("titre")
		contenu := r.Form.Get("contenu")

		_, err = db.Exec("INSERT INTO topics (utilisateurs_pseudo, utilisateur_id, titre, contenu) VALUES (?, ?, ?, ?)", pseudo, id, titre, contenu)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tpl := template.Must(template.ParseFiles("pages/create_topic.html"))
	tpl.Execute(w, nil)
}


func viewAllTopicsPage(w http.ResponseWriter, r *http.Request) {
	topics, err := getAllTopics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tpl := template.Must(template.ParseFiles("pages/all_topics.html"))
	tpl.Execute(w, topics)
}

func viewTopicDetailsPage(w http.ResponseWriter, r *http.Request) {
	// Récupération de l'ID du sujet à partir de la requête
	topicID := r.URL.Query().Get("id")
	
	// Conversion de topicID en entier
	id, err := strconv.Atoi(topicID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération des détails du sujet depuis la base de données
	topic, err := getTopicDetails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération des commentaires associés à ce sujet
	comments, err := getCommentsForTopic(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Affichage des détails du sujet et des commentaires sur la page
	tpl := template.Must(template.ParseFiles("pages/topic_details.html"))
	tpl.Execute(w, struct {
		Topic    Topic
		Comments []Commentaire
	}{
		Topic:    topic,
		Comments: comments,
	})
}


func getTopicDetails(topicID int) (Topic, error) {
	var topic Topic

	err := db.QueryRow("SELECT id, utilisateur_id, utilisateurs_pseudo, titre, contenu FROM topics WHERE id = ?", topicID).Scan(&topic.ID, &topic.UtilisateurID, &topic.Utilisateur, &topic.Titre, &topic.Contenu)
	if err != nil {
		return Topic{}, err
	}

	return topic, nil
}


func getCommentsForTopic(topicID int) ([]Commentaire, error) {
	var comments []Commentaire

	rows, err := db.Query("SELECT id, utilisateur_id, topic_id, contenu FROM commentaires WHERE topic_id = ?", topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Commentaire
		err := rows.Scan(&comment.ID, &comment.UtilisateurID, &comment.TopicID, &comment.Contenu)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
