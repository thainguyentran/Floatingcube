package main_test

import (
	"atlas"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	main "atlas/cmd/quickbookweb"
	"atlas/cmd/server"

	"github.com/julienschmidt/httprouter"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

var app *main.App

var user1 = &atlas.QBUser{
	ID:           1,
	Email:        "hochiminh@communist.com",
	Name:         "Mr. Ho",
	Password:     "hoisme",
	IsActive:     true,
	IsSuperAdmin: true,
	DateCreated:  time.Now(),
	DateUpdated:  time.Now(),
}
var org1 = atlas.QBOrg{
	ID:                 1,
	Name:               "Floating Cube Studios",
	QBID:               123,
	QBDepositAccountID: "321",
	QBCompanyID:        "193514527926034",
	QBCredToken:        "qyprdLazLx1P4ilJuXLKAyRKmWSK10Ol7d62xdPNfTefkbum",
	QBCredSecret:       "RDFyX1zqZcelY3iCq5JFsD9Z1s2wB4iUvas3MiHi",
	QBWebHookToken:     "4e4c499b-8854-46fc-a1df-f7519abcb745",
}
var shop1 = atlas.QBShop{
	ID:             1,
	Name:           "FCS HCM",
	QBDepartmentID: 1,
}
var skipProjectFlag = flag.String("skipTest", "", "Skip the given test function")

type MockLogger struct{}

func (ml *MockLogger) Log(str string, v ...interface{}) {
	fmt.Printf("mockLogger: "+str+"\n", v...)
}

type HandleTester func(method string, params url.Values) *httptest.ResponseRecorder
type HandleBodyTester func(method string, body io.Reader) *httptest.ResponseRecorder

// Given the current test runner and an http.Handler, generate a
// HandleTester which will test its given input against the
// handler.
func GenerateHandleTester(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
) HandleTester {
	return GenerateHandleTesterWithURLParams(
		t,
		handleFunc,
		loggedIn,
		httprouter.Params{},
	)
}

// GenerateHandleTesterWithURLParams returns a HandleTester
// given a httprouter.Params
func GenerateHandleTesterWithURLParams(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	httpRouterParams httprouter.Params,
) HandleTester {
	// Given a method type ("GET", "POST", etc) and
	// parameters, serve the response against the handler and
	// return the ResponseRecorder.
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", strings.NewReader(params.Encode()))
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		ctx := context.WithValue(req.Context(), server.Params, httpRouterParams)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
			ctx = context.WithValue(ctx, server.SessionKeyName, "abcd1234")
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

// GenerateHandleBodyTesterWithHeaders returns a HandleBodyTester
// given header params
func GenerateHandleBodyTesterWithHeaders(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	httpRouterParams httprouter.Params,
	headerVars map[string]string,
) HandleBodyTester {
	return func(method string, body io.Reader) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", body)
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		for k, v := range headerVars {
			req.Header.Set(k, v)
		}
		ctx := context.WithValue(req.Context(), server.Params, httpRouterParams)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1.ID)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

// GenerateHandleTesterWithForm returns a HandleTester
// given a form
func GenerateHandleTesterWithForm(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	formVars map[string]string,
) HandleTester {
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", strings.NewReader(params.Encode()))
		ok(t, err)
		req.Form = url.Values{}
		for k, v := range formVars {
			req.Form.Add(k, v)
		}
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		ctx := context.WithValue(req.Context(), server.Params, params)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1.ID)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

// GenerateHandleBodyTesterWithURLParams returns a HandleBodyTester
// given a httprouter.Params
func GenerateHandleBodyTesterWithURLParams(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	httpRouterParams httprouter.Params,
) HandleBodyTester {
	return func(method string, body io.Reader) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", body)
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		ctx := context.WithValue(req.Context(), server.Params, httpRouterParams)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1.ID)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

// GenerateHandleBodyTesterWithHeaderAndBody returns a HandleBodyTester
// given a httprouter.Params
func GenerateHandleBodyTesterWithHeaderAndBody(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	httpRouterParams httprouter.Params,
	headerVars map[string]string,
) HandleBodyTester {
	return func(method string, body io.Reader) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", body)
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		for k, v := range headerVars {
			req.Header.Set(k, v)
		}
		ctx := context.WithValue(req.Context(), server.Params, httpRouterParams)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1.ID)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

func GenerateHandleTesterWithHeaders(
	t *testing.T,
	handleFunc http.Handler,
	loggedIn bool,
	httpRouterParams httprouter.Params,
	headerVars map[string]string,
	queryVars map[string]string,
) HandleTester {
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", strings.NewReader(params.Encode()))
		ok(t, err)
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		for k, v := range headerVars {
			req.Header.Set(k, v)
		}
		rawQuery := ""
		for k, v := range queryVars {
			rawQuery += k + "=" + v + "&"
		}
		if rawQuery != "" {
			rawQuery = rawQuery[:len(rawQuery)-1]
		}

		req.URL.RawQuery = rawQuery

		ctx := context.WithValue(req.Context(), server.Params, httpRouterParams)
		w := httptest.NewRecorder()
		if loggedIn {
			ctx = context.WithValue(ctx, server.UserKeyName, user1.ID)
			ctx = context.WithValue(ctx, server.OrgKeyName, org1.ID)
			ctx = context.WithValue(ctx, server.ShopKeyName, shop1.ID)
		}
		req = req.WithContext(ctx)
		handleFunc.ServeHTTP(w, req)
		return w
	}
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// skip used for skip the function with parameters
func skip(t *testing.T, flagName *string, skipName string) {
	if flagName == nil || *flagName == "" {
		return
	}

	if strings.Contains(skipName, *flagName) {
		t.SkipNow()
	}

	return
}

func TestMain(m *testing.M) {
	flag.Parse()
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %s", err)
	}
	r := server.NewRouter()
	ml := &MockLogger{}
	err = LoadQuickBookConfiguration(pwd)
	if err != nil {
		log.Printf("error loading configuration file: %s", err)
	}
	templatePath := path.Join(viper.GetString("path"), "templates")
	app = main.SetupApp(r, ml, []byte("some-secret"), templatePath)

	retCode := m.Run()
	os.Exit(retCode)
}

func LoadQuickBookConfiguration(pwd string) error {
	viper.SetConfigName("dev_test_config")
	//viper.AddConfigPath(pwd)
	_, devPath, _, _ := runtime.Caller(1)
	devPath = path.Dir(devPath)
	viper.AddConfigPath(devPath)
	viper.SetDefault("path", devPath)
	return viper.ReadInConfig() // Find and read the config file
}
