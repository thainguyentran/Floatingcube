package main

import (
	"atlas"
	"atlas/cmd/server"
	"net/http"
)

// GetAuthAPIHandler is for checking if a token is valid, it returns proper API responses in the body.
func (a *App) GetAuthAPIHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		a.Rndr.JSON(w, 200, server.NewAPIResponse(200, "Token is valid"))
		return nil
	}
}

// HeadAuthAPIHandler is for checking if a token is valid, but using HEAD.
func (a *App) HeadAuthAPIHandler(db atlas.AtlasSessionDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		tokenString := req.Header.Get("Authorization")

		if tokenString == "" {
			a.Rndr.JSON(w, 403, nil)
			return nil
		}

		_, err := db.GetAtlasSessionByToken(tokenString)
		if err != nil {
			a.Rndr.JSON(w, 401, nil)
			return nil
		}

		a.Rndr.JSON(w, 200, nil)
		return nil
	}
}