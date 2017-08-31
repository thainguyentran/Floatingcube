package main

import (
	"atlas"
	"atlas/cmd/server"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
)

const sessionName = "session"
const userKeyName = "user"
const sessionKeyName = "session_key"

// LoginPageHandler is the handler for displaying the login page.
func (a *App) LoginPageHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		u, _ := getUser(req)
		// user is logged in already
		if u != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return nil
		}
		lp := &localPresenter{
			PageTitle:       "Login",
			PageURL:         "/login",
			GlobalPresenter: a.Gp,
		}
		a.Rndr.HTML(w, http.StatusOK, "login", lp)
		return nil
	}
}

// LoginPostHandler is the handler for dealing with user login input.
func (a *App) LoginPostHandler(db atlas.QBUserDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		u, _ := getUser(req)
		// user is logged in already
		if u != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return nil
		}
		email, pass := strings.TrimSpace(req.FormValue("email")), req.FormValue("password")
		// TODO: add proper flashes and validation
		if email == "" || pass == "" {
			http.Redirect(w, req, "/login", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "validation error: email or password cannot be empty", fmt.Errorf("email or password cannot be empty"))
		}
		if !govalidator.IsEmail(email) {
			http.Redirect(w, req, "/login", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "email has to be a valid email address", fmt.Errorf("email has to be a valid email address"))

		}
		u, err := db.GetQBUserByEmail(email)
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusFound)
			return server.New500Error("internal server error: unable to retrieve user by email", err)
		}
		if !u.IsCorrectPassword(pass, db) {
			http.Redirect(w, req, "/login", http.StatusFound)
		}

		sess, err := db.CreateAtlasWebSession(u.ID)
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusFound)
			return server.New500Error("internal server error: error during create web session", err)
		}

		session, err := a.Store.Get(req, sessionName)
		if err != nil {
			return server.New500Error("internal server error: error during getting of session", err)
		}
		session.Values[sessionKeyName] = sess.SessionKey
		session.Save(req, w)
		http.Redirect(w, req, "/w", http.StatusFound)
		return nil
	}
}

// LogoutHandler is the handler for logging out.
// TODO: set the flashes properly when redirecting.
func (a *App) LogoutHandler(db atlas.QBUserDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		ssk, err := getSessionKey(req)
		if err != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return server.New500Error("internal server error: error retrieving sessionKey from request", err)
		}
		err = db.DeleteAtlasWebSession(ssk)
		if err != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return server.New500Error("internal server error: error deleting session from db", err)
		}
		session, err := a.Store.Get(req, sessionName)
		if err != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return server.New500Error("internal server error: error during getting of session", err)
		}
		delete(session.Values, sessionKeyName)
		session.Save(req, w)
		http.Redirect(w, req, "/", http.StatusFound)
		return nil
	}
}

// UserIndexManagementHandler displays the users management page.
func (a *App) UserIndexManagementHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

// UserInfoEditPageHandler is the handler for displaying the edit form for one user.
func (a *App) UserInfoEditPageHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}

// UserInfoEditPostHandler is the handler for handling the Post user data.
func (a *App) UserInfoEditPostHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		return nil
	}
}
