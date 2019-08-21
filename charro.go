package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	_ "github.com/lib/pq"
	"github.com/nerijus-st/charro/lastfm"
	"github.com/nerijus-st/charro/spotify"
	spotifyWrapper "github.com/zmb3/spotify"
)

var (
	auth = spotifyWrapper.NewAuthenticator(redirectURI,
		spotifyWrapper.ScopeUserReadPrivate,
		spotifyWrapper.ScopePlaylistModifyPrivate,
		spotifyWrapper.ScopeUserTopRead,
	)
	ch          = make(chan *spotifyWrapper.Client)
	state       = "zxvcasdfqw"
	client      *spotifyWrapper.Client
	redirectURI = os.Getenv("OAUTH2_REDIRECT_URI")
	pingChan    = make(chan string)
)

//Data to pass to templates
type Data struct {
	User          *spotifyWrapper.PrivateUser
	URL           string
	LastFMTracks  *lastfm.Tracks
	SpotifyTracks *spotify.Tracks
	PlaylistID    spotifyWrapper.ID
	Success       bool
	Error         string
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	client := auth.NewClient(tok)
	ch <- &client

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	logToDB("users", user.User.DisplayName, "")

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	log.Println("Got request for:", r.URL.String())

	if client == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/main.html"))
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		data := Data{
			User: user,
		}
		tmpl.ExecuteTemplate(w, "base", data)
	}

}

func lastFMHandle(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/lastfm.html"))
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}

		if r.Method != http.MethodPost {
			data := Data{
				User: user,
			}
			tmpl.ExecuteTemplate(w, "base", data)
		} else {
			LastFMForm := lastfm.Form{
				Username: r.FormValue("lastfm-username"),
				Period:   r.FormValue("period"),
				Limit:    r.FormValue("limit"),
			}

			lastFMTracks, err := lastfm.GetTopTracks(&LastFMForm.Username, &LastFMForm.Period, &LastFMForm.Limit)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			spotifyTrackIDs, err := spotify.GetTracksBasedOnLastFM(client, lastFMTracks)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			playlist, err := spotify.GeneratePlaylist(client, spotifyTrackIDs, &LastFMForm.Period, &LastFMForm.Limit)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			data := Data{
				User:         user,
				LastFMTracks: lastFMTracks,
				PlaylistID:   playlist.ID,
				Success:      true,
			}
			tmpl.ExecuteTemplate(w, "base", data)
		}
	}
}

func logError(w *http.ResponseWriter, err error, userDisplayName string, tmpl *template.Template) {
	fmt.Println(err)
	logToDB("errors", userDisplayName, err.Error())
	data := Data{Error: err.Error()}
	tmpl.ExecuteTemplate(*w, "base", data)
}

func spotifyHandle(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/spotify.html"))
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}

		if r.Method != http.MethodPost {
			data := Data{
				User: user,
			}
			tmpl.ExecuteTemplate(w, "base", data)
		} else {

			SpotifyForm := spotify.Form{
				Limit:     r.FormValue("limit"),
				TimeRange: r.FormValue("time_range"),
			}

			limitInt, err := strconv.Atoi(SpotifyForm.Limit)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			spotifyTrackIDs, err := spotify.GetTopTracks(client, SpotifyForm.TimeRange, limitInt)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			playlist, err := spotify.GeneratePlaylist(client, spotifyTrackIDs, &SpotifyForm.TimeRange, &SpotifyForm.Limit)
			if err != nil {
				logError(&w, err, user.User.DisplayName, tmpl)
				return
			}

			data := Data{
				User:       user,
				PlaylistID: playlist.ID,
				Success:    true,
			}
			tmpl.ExecuteTemplate(w, "base", data)
		}
	}
}

func loginHandle(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
		loginURL := auth.AuthURL(state)
		data := Data{
			URL: loginURL,
		}
		tmpl.ExecuteTemplate(w, "base", data)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logToDB(logType, user, errorDesc string) {
	database, err := connect()

	if err != nil {
		fmt.Printf("[x] Warning: could not connect to database. Logging is off. %s\n", err.Error())
		return
	}

	if logType == "users" {
		_, err = database.Query(fmt.Sprintf("INSERT INTO %s (username) VALUES ('%s');", logType, user))
		if err != nil {
			fmt.Printf(fmt.Sprintf("INSERT INTO %s (username) VALUES ('%s');", logType, user))
			fmt.Printf("[x] Error on insert. Reason: %s\n", err.Error())
		}
	} else {
		_, err = database.Query(fmt.Sprintf("INSERT INTO %s (username, error_desc) VALUES ('%s', '%s');", logType, user, errorDesc))
		if err != nil {
			fmt.Printf(fmt.Sprintf("INSERT INTO %s (username, error_desc) VALUES ('%s', '%s');", logType, user, errorDesc))
			fmt.Printf("[x] Error on insert. Reason: %s\n", err.Error())
		}
	}

	defer database.Close()
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func connect() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	return database, nil
}

func createTables() {
	database, err := connect()
	if err != nil {
		fmt.Printf("[x] Warning: could not connect to database. Logging is off. %s\n", err.Error())
	} else {
		_, err = database.Query("CREATE TABLE IF NOT EXISTS USERS (id SERIAL, username varchar(255), logged_at TIMESTAMPTZ NOT NULL DEFAULT NOW());")
		if err != nil {
			fmt.Printf("[x] Error on creating users table. Reason: %s\n", err.Error())
		}
		_, err = database.Query("CREATE TABLE IF NOT EXISTS ERRORS (id SERIAL, username varchar(255),error_desc varchar(255), errored_at TIMESTAMPTZ NOT NULL DEFAULT NOW());")
		if err != nil {
			fmt.Printf("[x] Error on creating users table. Reason: %s\n", err.Error())
		}

		defer database.Close()
	}
}

func main() {

	createTables()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	http.HandleFunc("/", mainHandle)
	http.HandleFunc("/auth", completeAuth)
	http.HandleFunc("/login", loginHandle)
	http.HandleFunc("/lastfm", lastFMHandle)
	http.HandleFunc("/spotify", spotifyHandle)
	http.HandleFunc("/favicon.ico", faviconHandler)

	go func() {
		client = <-ch
	}()

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}
