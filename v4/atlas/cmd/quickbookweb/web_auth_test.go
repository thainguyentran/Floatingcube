package main_test

import (
		"atlas"
		"testing"
		"fmt"
		"net/url"
		"net/http"

		"golang.org/x/crypto/bcrypt"
)
type MockQBUserDB struct {
	hasError		bool
	mockhashedPassword	[]byte
	mockTx			*atlas.Tx
}

func (db *MockQBUserDB) Begin() (*atlas.Tx, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	
	return db.mockTx, nil
}

func (db *MockQBUserDB) Rollback(tx *atlas.Tx) error {
	if db.hasError {
		return fmt.Errorf("some error")
	}
	return tx.Rollback()
}

func (db *MockQBUserDB) CreateQBUser(u atlas.QBUser) (*atlas.QBUser, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	return user1, nil
}

func (db *MockQBUserDB) GetQBUserByEmail(email string) (*atlas.QBUser, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	return user1, nil
}

func (db *MockQBUserDB) GetPassword(userID int) ([]byte, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user1.Password), 13)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func (db *MockQBUserDB) GetQBUserByID(userID int) (*atlas.QBUser, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	return user1, nil
}

func (db *MockQBUserDB) DeleteAtlasWebSession(sessionKey string) error {
	if db.hasError {
		return fmt.Errorf("some error")
	}
	return nil
}

func (db *MockQBUserDB) CreateAtlasWebSession(userID int) (*atlas.WebSession, error) {
	if db.hasError {
		return nil, fmt.Errorf("some error")
	}
	aws := &atlas.WebSession{
		UserID:     userID,
		SessionKey: "w1232445",
	}
	return aws, nil
}

func TestLoginPageHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	lp := app.LoginPageHandler()
	//log in fail
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected signup page to return 200 instead got %d", w.Code)
	//log in success
	test = GenerateHandleTester(t, app.Wrap(lp), true)
	w = test("GET", url.Values{})
	assert(t, w.Code == http.StatusFound, "expected signup page to return 304 instead got %d", w.Code)
}

func TestLoginPostHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	mockDB := &MockQBUserDB{
		hasError:			false,
	}
	lp := app.LoginPostHandler(mockDB)

	test := GenerateHandleTester(t, app.Wrap(lp), false)
	//correct password
	w := test("POST", url.Values{"email": {"hochiminh@communist.com"}, "password": {"hoisme"}})
	assert(t, w.Code == http.StatusFound, "expected successful login with proper POST inputs to redirect 302 instead got %d", w.Code)
	assert(t, len(w.HeaderMap["Set-Cookie"]) > 0, "expected session cookie to be set on successful login.")

	// Wrong password
	w = test("POST", url.Values{"email": {"hochiminh@communist.com"}, "password": {"tranisme"}})
	assert(t, w.Code == http.StatusFound, "expected successful login with proper POST inputs to redirect 302 instead got %d", w.Code)
}

func TestLogoutHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	mockDB := &MockQBUserDB{
		hasError:			false,
	}
	lp := app.LogoutHandler(mockDB)
	test := GenerateHandleTester(t, app.Wrap(lp), true)
	w := test("POST", url.Values{})
	assert(t, w.Code == http.StatusFound, "expected successful logout to redirect 302 instead got %d", w.Code)
	assert(t,
		len(w.HeaderMap["Location"]) > 0 && w.HeaderMap["Location"][0] == "/",
		"expected redirect location on successful logout to be / instead got %s", w.HeaderMap["Location"])
	
}

func TestUserIndexManagementHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	lp := app.UserIndexManagementHandler()
	
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected page to return 200 instead got %d", w.Code)
}

func TestUserInfoEditPageHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	lp := app.UserInfoEditPageHandler()
	
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected page to return 200 instead got %d", w.Code)
}

func TestUserInfoEditPostHandler(t *testing.T) {
	skip(t, skipProjectFlag, "quickbook")
	lp := app.UserInfoEditPostHandler()
	
	test := GenerateHandleTester(t, app.Wrap(lp), false)
	w := test("GET", url.Values{})
	assert(t, w.Code == http.StatusOK, "expected page to return 200 instead got %d", w.Code)
}
