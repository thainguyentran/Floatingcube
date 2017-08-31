package main

import (
	"atlas"
	"atlas/cmd/server"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/garyburd/go-oauth/oauth"
)

// WebStartPageHandler is the handler to select Quickbooks or Odoo
func (a *App) WebStartPageHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		lp := &localPresenter{
			PageTitle:       "Setup",
			PageURL:         "/start",
			GlobalPresenter: a.Gp,
		}
		switch req.Method {
		case "GET":
			a.Rndr.HTML(w, http.StatusOK, "start1", lp)
			return nil
		case "POST":
			qb := strings.TrimSpace(req.FormValue("quickbooks"))
			if qb == "on" {
				http.Redirect(w, req, "/start/2", http.StatusFound)
			}
			return nil
		}
		return nil
	}
}

// WebStart2PageHandler is the handler to display after the user has selected Quickbooks. In this case we get the user to create a superadmin user
func (a *App) WebStart2PageHandler() server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		if u, _ := getUser(req); u != nil {
			http.Redirect(w, req, "/", http.StatusFound)
			return nil
		}
		fs := a.getFlashes(w, req)
		p := &struct {
			Flashes []interface{}
			localPresenter
		}{
			Flashes:        fs,
			localPresenter: localPresenter{
				PageTitle: "Set up superadmin acount", 
				PageURL: "/start/2", 
				GlobalPresenter: a.Gp}}
		err := a.Rndr.HTML(w, http.StatusOK, "start2", p)
		if err != nil {
			return err
		}
		return nil
	}
}

// WebStart2PostHandler is the handler to handle the post request from WebStart2PageHandler. It creates a superadmin
func (a *App) WebStart2PostHandler(db atlas.QBUserDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		email, pass := strings.TrimSpace(req.FormValue("email")), req.FormValue("password")
		if email == "" || pass == "" {
			a.saveFlash(w, req, "You cannot leave email or password empty")
			http.Redirect(w, req, "/start/2", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "No email/password provided", nil)
		}
		if !govalidator.IsEmail(email) {
			a.saveFlash(w, req, "email has to be a valid email address")
			http.Redirect(w, req, "/start/2", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "incorrect email address", nil)
		}

		// TODO: handle the case where user email already exist
		user, err := db.CreateQBUser(atlas.QBUser{Email: email, Password: pass, IsSuperAdmin: true})
		if err != nil {
			return server.New500Error("internal server error: something went wrong when creating user", err)
		}
		sess, err := db.CreateAtlasWebSession(user.ID)
		if err != nil {
			return server.New500Error("internal server error: error during create web session", err)
		}

		session, err := a.Store.Get(req, sessionName)
		if err != nil {
			return server.New500Error("internal server error: error during saving of session", err)
		}
		session.Values[sessionKeyName] = sess.SessionKey
		session.Save(req, w)
		http.Redirect(w, req, "/start/3", http.StatusFound)
		return nil
	}
}

// WebStart3PageHandler is the handler to display after the user has created a superadmin. In this case we get the user to create an org.
func (a *App) WebStart3PageHandler(db atlas.QBOrgDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		u, err := getUser(req)
		if err != nil {
			return server.New500Error("error retrieving user from request", err)
		}
		_, err = db.IncompleteGetAllQBOrgForUser(u.ID)
		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return nil
		}

		lp := &localPresenter{
			PageTitle:       "Setup Organisation",
			PageURL:         "/start/3",
			User:            u,
			GlobalPresenter: a.Gp,
		}
		a.Rndr.HTML(w, http.StatusOK, "start3", lp)
		return nil
	}
}

// WebStart3PostHandler is the handler that creates an org and redirects to the shop setup page.
// TODO add proper flashes and validation warnings
func (a *App) WebStart3PostHandler(db atlas.QBWebStart3DB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		name := strings.TrimSpace(req.FormValue("name"))
		if name == "" {
			http.Redirect(w, req, "/start/3", http.StatusFound)
			return nil
		}
		// Setting up org, paymentmethods
		org := atlas.QBOrg{Name: name}
		newOrg, err := db.CreateQBOrg(org)
		if err != nil {
			return server.New500Error("error while creating organisation", err)
		}
		paymentMethods := []*atlas.QBPaymentMethod{
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Cash (SGD)",
				DisplayName: "Cash (SGD)",
				Code:        "cash",
				Type:        "NON_CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Cash Card",
				DisplayName: "Cash Card",
				Code:        "ccard",
				Type:        "NON_CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Credit Card (SGD)",
				DisplayName: "Credit Card (SGD)",
				Code:        "crc",
				Type:        "CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "EZ Link",
				DisplayName: "EZ Link",
				Code:        "ezlin",
				Type:        "NON_CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Nets",
				DisplayName: "Nets",
				Code:        "nets",
				Type:        "NON_CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Voucher",
				DisplayName: "Voucher",
				Code:        "vcher",
				Type:        "NON_CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Visa",
				DisplayName: "Visa",
				Code:        "visa",
				Type:        "CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Master",
				DisplayName: "Master",
				Code:        "master",
				Type:        "CREDIT_CARD",
			},
			&atlas.QBPaymentMethod{
				OrgID:       newOrg.ID,
				Name:        "Amex",
				DisplayName: "American Express",
				Code:        "amex",
				Type:        "CREDIT_CARD",
			},
		}
		for _, pm := range paymentMethods {
			_, err = db.CreateQBPaymentMethod(*pm)
			if err != nil {
				return server.New500Error("error while creating payment methods for organisation", err)
			}
		}

		http.Redirect(w, req, "/start/4", http.StatusFound)
		return nil
	}
}

// WebStart4PageHandler is the handler for connecting with Quickbooks.
func (a *App) WebStart4PageHandler(db atlas.QBOrgDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		u, err := getUser(req)
		if err != nil {
			return server.New500Error("cannot get user from context", err)
		}

		orgs, err := db.IncompleteGetAllQBOrgForUser(u.ID)
		if err != nil {
			return server.New500Error("error retrieving orgs for user", err)
		}
		pp := struct {
			Orgs []*atlas.QBOrg
			*localPresenter
		}{
			Orgs: orgs,
			localPresenter: &localPresenter{
				PageTitle:       "Connect to Quickbooks",
				PageURL:         "/start/4",
				GlobalPresenter: a.Gp,
				User:            u,
			},
		}
		a.Rndr.HTML(w, http.StatusOK, "start4", pp)
		return nil
	}
}

// QuickbooksCallback is the callback endpoint for Quickbooks after the auth dance.
func (a *App) QuickbooksCallback(db atlas.QBOrgDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		sess, err := a.Store.Get(req, tempCredName)
		if err != nil {
			return server.New500Error("error grabbing temp credentials", err)
		}
		var tempCred = &oauth.Credentials{}
		tempCred, ok := sess.Values[tempCredKey].(*oauth.Credentials)
		if !ok {
			return server.New500Error("error casting oauth.Credentials", err)
		}

		if tempCred.Token != req.FormValue("oauth_token") {
			return server.New500Error("unknown oauth_token", fmt.Errorf("unknown oauth_token in request"))
		}
		tokenCred, _, err := a.oauthClient.RequestToken(nil, tempCred, req.FormValue("oauth_verifier"))
		if err != nil {
			return server.New500Error("error getting request token", err)
		}
		val := sess.Values[tempOrgIDKey]
		orgID, ok := val.(int)
		if !ok {
			return server.New500Error("unable to type cast orgID from session", fmt.Errorf("failure to typecast orgID in request session"))
		}
		delete(sess.Values, tempCredKey)
		sess.Save(req, w)

		org, err := db.GetQBOrg(orgID)
		if err != nil {
			return server.New500Error("unable to retrieve org", err)
		}

		org.QBCompanyID = req.FormValue("realmId")
		org.QBCredSecret = tokenCred.Secret
		org.QBCredToken = tokenCred.Token
		_, err = db.UpdateQBOrg(*org)
		if err != nil {
			return server.New500Error("error saving org", err)
		}

		http.Redirect(w, req, "/start/5", http.StatusFound)
		return nil
	}
}

// WebStart5PageHandler is the handler to display orgs, for creating a shop each.
func (a *App) WebStart5PageHandler(db atlas.QBOrgShopDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		u, err := getUser(req)
		if err != nil {
			return server.New500Error("error retrieving user from request", err)
		}
		orgs, err := db.IncompleteGetAllQBOrgForUser(u.ID)
		if err != nil {
			return server.New500Error("error retrieving orgs", err)
		}

		if len(orgs) == 0 {
			http.Redirect(w, req, "/start/3", http.StatusFound)
			return nil
		}

		// if any shops exist for any of the orgs, we get out of the setup flow
		for _, o := range orgs {
			shops, err := db.GetAllShopsForOrg(o.ID)
			if err != nil {
				a.Logr.Log("error retrieving shops for org ", o.Name, err)
				continue
			}
			if len(shops) > 0 {
				http.Redirect(w, req, "/w", http.StatusFound)
				return nil
			}
		}

		p := struct {
			Orgs []*atlas.QBOrg
			*localPresenter
		}{
			Orgs: orgs,
			localPresenter: &localPresenter{
				PageTitle:       "Setup Shop",
				PageURL:         "/start/4",
				User:            u,
				GlobalPresenter: a.Gp,
			}}
		a.Rndr.HTML(w, http.StatusOK, "start5", p)
		return nil
	}
}

// WebStart5PostHandler is the handler for creating a shop. It then redirects to the home page.
func (a *App) WebStart5PostHandler(db atlas.QBOrgShopSessionDB) server.HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		name, orgIDString := strings.TrimSpace(req.FormValue("name")), req.FormValue("orgid")
		if name == "" || orgIDString == "" {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "name or org id cannot be empty", fmt.Errorf("bad input for creating shop: either orgID or name is empty"))
		}

		orgID, err := strconv.Atoi(orgIDString)
		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.NewError(http.StatusBadRequest, "error converting orgID from request", err)
		}

		org, err := db.GetQBOrg(orgID)
		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.New500Error("error retrieving org", err)
		}

		shop, err := db.CreateQBShop(atlas.QBShop{Name: name, OrgID: orgID})
		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.New500Error("error saving shop", err)
		}

		u, err := getUser(req)
		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.New500Error("error getting user from request", err)
		}

		_, err = db.CreateAtlasSession(atlas.AtlasSession{
			UserID:   u.ID,
			UserName: u.Name,
			OrgID:    org.ID,
			OrgName:  org.Name,
			ShopID:   shop.ID,
			ShopName: shop.Name,
		})

		if err != nil {
			http.Redirect(w, req, "/start/4", http.StatusFound)
			return server.New500Error("error creating new atlas session", err)
		}

		http.Redirect(w, req, "/w", http.StatusFound)
		return nil
	}
}
