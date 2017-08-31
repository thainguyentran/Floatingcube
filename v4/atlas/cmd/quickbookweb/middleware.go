package main

import (
	"atlas"
	"atlas/cmd/server"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	tempCredName = "tempCredName-quickbook"
	tempCredKey  = "tempCred-quickbook"
	tokenCredKey = "tokenCred-quickbook"
	companyKey   = "company"

)

// ifModifiedSinceMiddleware is middleware wrapper to protec tauthentication endpoint.
func (a *App) ifModifiedSinceMiddleware(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, req *http.Request) error {
		modifiedSince := req.Header.Get(ifModifiedSince)
		if modifiedSince != "" {
			psqlTime, err := ConvertRFCDatetime2PsqlDatetime(modifiedSince)
			if err != nil {
				a.Logr.Log("error on converting from RFC datetime to Psql datetime, %s", err)
			}
			ctx := context.WithValue(req.Context(), server.IfModifiedSince, psqlTime)
			req = req.WithContext(ctx)
		}
		next.ServeHTTP(rw, req)
		return nil
	}
	return a.Wrap(fn)
}

// authAtlasMiddleware is the middleware for authenticate request calls from V4
func (a *App) authAtlasMiddleware(db atlas.AtlasSessionDB) func(http http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, req *http.Request) error {
			tokenString := req.Header.Get("Authorization")

			if tokenString == "" {
				return server.NewAPIError(403, "You need to provide a token key to use this endpoint", nil)
			}

			session, err := db.GetAtlasSessionByToken(tokenString)
			if err != nil {
				return server.NewAPIError(401, "Bad credentials", err)
			}
			ctx := context.WithValue(req.Context(), server.UserKeyName, session.UserID)
			ctx = context.WithValue(ctx, server.ShopKeyName, session.ShopID)
			req = req.WithContext(context.WithValue(ctx, server.OrgKeyName, session.OrgID))
			next.ServeHTTP(rw, req)
			return nil
		}
		return a.Wrap(fn)
	}
}

// webUserAtlasMiddleware detects and adds the user object to a web request context
// some errors are not logged because it's normal for some requests to not have users or sessions. (e.g. non logged-in requests)
func (a *App) webUserAtlasMiddleware(db atlas.AtlasSessionDB) func(http http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			session, err := a.Store.Get(req, sessionName)
			if err != nil {
				next.ServeHTTP(w, req)
				return
			}

			sessionKey, ok := session.Values[sessionKeyName]
			if !ok {
				next.ServeHTTP(w, req)
				return
			}

			ssk, ok := sessionKey.(string)
			if !ok {
				a.Logr.Log("error converting sessionKey into string for request %+v", req)
				next.ServeHTTP(w, req)
				return
			}

			u, err := db.GetUserFromSession(ssk)
			if err != nil {
				a.Logr.Log("error getting user from session key %s: %s", ssk, err)
				delete(session.Values, sessionKey)
				session.Save(req, w)
				next.ServeHTTP(w, req)
				return
			}

			ctx := context.WithValue(req.Context(), userKeyName, u)
			ctx = context.WithValue(ctx, sessionKeyName, ssk)
			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
			return
		}
		return http.HandlerFunc(fn)
	}
}

// webHookAuthMiddleware verify the token from developer.intuit.com is correct
func (a *App) webHookAuthMiddleware(db atlas.QBOrgWebHookDB) func(http http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			jsonBody, err := ioutil.ReadAll(req.Body)
			if err != nil {
				a.Logr.Log("error reading webhook payload: %s", err)
				http.Error(w, "webhook payload in bad form", 400)
				return
			}
			rdr1 := ioutil.NopCloser(bytes.NewBuffer(jsonBody))
			rdr2 := ioutil.NopCloser(bytes.NewBuffer(jsonBody))

			var webPayload atlas.WebPayload
			err = json.NewDecoder(rdr1).Decode(&webPayload)
			if err != nil {
				a.Logr.Log("error reading webhook payload: %s", err)
				http.Error(w, "webhook payload in bad form", 400)
				return
			}
			if len(webPayload.EventNotification) < 1 {
				http.Error(w, "empty payload", 400)
				return
			}

			companyID := webPayload.EventNotification[0].RealmID
			qbOrg, err := db.GetQBOrgByCompanyID(companyID)
			if err != nil {
				a.Logr.Log("no company with id = %s exist", companyID)
				http.Error(w, "company not found", 400)
				return
			}

			signedBody := req.Header.Get(intuitSignature)
			if !CheckMAC(jsonBody, signedBody, qbOrg.QBWebHookToken) && a.IsProduction {
				http.Error(w, "Not logged in.", 403)
				return
			}

			req.Body = rdr2
			ctx := context.WithValue(req.Context(), server.OrgKeyName, qbOrg.ID)
			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
			return
		}
		return http.HandlerFunc(fn)
	}
}

// TODO: do permissions later la
// webAuthMiddleware blocks access to the webpages from un-logged-in users
func (a *App) webAuthMiddleware(db atlas.AtlasSessionDB) func(http http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, req *http.Request) {
			user, err := getUser(req)
			if err != nil {
				if err != ErrNotLoggedIn {
					a.Logr.Log("middleware error: %s", err)
				}
				http.Redirect(w, req, "/login", http.StatusFound)
				return
			}

			// Always serve if is superadmin
			if user.IsSuperAdmin {
				next.ServeHTTP(w, req)
				return
			}

			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(fn)
	}
}

var ErrNotLoggedIn = fmt.Errorf("no user in context")

func getUser(req *http.Request) (*atlas.QBUser, error) {
	uu := req.Context().Value(server.UserKeyName)
	if uu == nil {
		return nil, ErrNotLoggedIn
	}
	u, ok := uu.(*atlas.QBUser)
	if !ok {
		return nil, fmt.Errorf("error converting user value to QBUser pointer")
	}
	return u, nil
}

func (a *App) getFlashes(w http.ResponseWriter, req *http.Request) []interface{} {
	session, _ := a.Store.Get(req, sessionName)
	fs := session.Flashes()

	session.Save(req, w)
	return fs
}

func (a *App) saveFlash(w http.ResponseWriter, req *http.Request, msg string) error {
	session, err := a.Store.Get(req, sessionName)
	if err != nil {
		return err
	}
	session.AddFlash(msg)
	err = session.Save(req, w)
	if err != nil {
		return err
	}
	return nil
}

func getSessionKey(req *http.Request) (string, error) {
	s := req.Context().Value(server.SessionKeyName)
	if s == nil {
		return "", fmt.Errorf("error retrieving session key from request")
	}
	ssk, ok := s.(string)
	if !ok {
		return "", fmt.Errorf("error converting session key to string")
	}
	return ssk, nil
}

// Throttler code from https://github.com/pressly/chi/blob/master/middleware/throttler.go
const (
	errCapacityExceeded = "Server capacity exceeded."
	errTimedOut         = "Timed out while waiting for a pending request to complete."
	errContextCanceled  = "Context was canceled."
)

// token represents a request that is being processed.
type token struct{}

// throttler limits number of currently processed requests at a time.
type throttler struct {
	h              http.Handler
	tokens         chan token
	backlogTokens  chan token
	backlogTimeout time.Duration
}

// ThrottleBacklog is a middleware that limits number of currently processed
// requests at a time and provides a backlog for holding a finite number of
// pending requests.
func (a *App) ThrottleBacklog(limit int, backlogLimit int, backlogTimeout time.Duration) func(http.Handler) http.Handler {
	if limit < 1 {
		a.Logr.Log("Throttle/middleware: Throttle expects limit > 0")
	}

	if backlogLimit < 0 {
		a.Logr.Log("Throttle/middleware: Throttle expects backlogLimit to be positive")
	}

	t := throttler{
		tokens:         make(chan token, limit),
		backlogTokens:  make(chan token, limit+backlogLimit),
		backlogTimeout: backlogTimeout,
	}

	// Filling tokens.
	for i := 0; i < limit+backlogLimit; i++ {
		if i < limit {
			t.tokens <- token{}
		}
		t.backlogTokens <- token{}
	}

	fn := func(h http.Handler) http.Handler {
		t.h = h
		return &t
	}

	return fn
}

// ServeHTTP is the primary throttler request handler
func (t *throttler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		time.Sleep(5 * time.Second)
		http.Error(w, errContextCanceled, http.StatusLocked)
		return
	case btok := <-t.backlogTokens:
		timer := time.NewTimer(t.backlogTimeout)

		defer func() {
			t.backlogTokens <- btok
		}()

		select {
		case <-timer.C:
			time.Sleep(3 * time.Second)
			http.Error(w, errTimedOut, http.StatusLocked)
			return
		case <-ctx.Done():
			time.Sleep(3 * time.Second)

			http.Error(w, errContextCanceled, http.StatusLocked)
			return
		case tok := <-t.tokens:
			defer func() {
				t.tokens <- tok
			}()
			t.h.ServeHTTP(w, r)
		}
		return
	default:
		time.Sleep(3 * time.Second)
		http.Error(w, errCapacityExceeded, http.StatusLocked)
		return
	}
}
