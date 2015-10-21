package studip

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// DefaultUserAgent if not already set by API.Header
var DefaultUserAgent = "StudIP Golang Library (github.com/blang/studip)"

// StudIPAPIBaseURL, the url to the api
var StudIPAPIBaseURL = "https://studip.uni-passau.de/studip/api.php/"

// HTTPClient is a general interface for http requests.
// Implemented by *http.Client
type HTTPClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// API to communicate with StudIP
type API struct {
	Header http.Header
	Client HTTPClient
}

// StatusCodeError is returned if the error is only specified by an invalid status code.
type StatusCodeError struct {
	Code int
	Msg  string
}

// Error returns the error message
func (e StatusCodeError) Error() string {
	return e.Msg
}

// APIError is returned if the remote api responds with an error.
type APIError struct {
	Msg          string
	InvalidLogin bool
	Parent       error
}

// Error returns the error message
func (e APIError) Error() string {
	return e.Msg
}

// DocumentTree represents a tree of all semesters and corresponding documents.
// An error is returned if the request could not be completed.
// The error can be a StatusCodeError.
func (a *API) DocumentTree() (*DocumentTree, error) {
	req, err := http.NewRequest("GET", StudIPAPIBaseURL+"/studip-client-core/documenttree/", nil)
	if err != nil {
		return nil, err
	}

	a.applyHeader(req)

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &StatusCodeError{
			Code: resp.StatusCode,
			Msg:  fmt.Sprintf("Invalid status code: %d", resp.StatusCode),
		}
	}

	dec := json.NewDecoder(resp.Body)
	var docTree DocumentTree
	err = dec.Decode(&docTree)
	if err != nil {
		return nil, err
	}
	return &docTree, nil
}

// applyHeader applies the api header and default user-agent.
func (a *API) applyHeader(req *http.Request) {
	if a.Header != nil {
		req.Header = a.Header
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Add("User-Agent", DefaultUserAgent)
	}
}

// SAML Response Form Regexp
var reAuth = regexp.MustCompile(`(?Us)form[\w\W]*action="([^"]+)"[\w\W]*input[\w\W]*name="([^"]+)"[\w\W]*value="([^"]+)"[\w\W]*input[\w\W]*name="([^"]+)"[\w\W]*value="([^"]+)"`)

// Login form Regexp
var reLoginForm = regexp.MustCompile(`(?Us)input[\w\W]*name="j_username"[\w\W]*input[\w\W]*name="j_password"`)

// Invalid login
var reInvalidLogin = regexp.MustCompile(`(?Us)loginerror-body[\w\W]*>([^<]+)<`)

// Urls used for studip login
const (
	loginURL          = "https://studip.uni-passau.de/studip/index.php?again=yes&sso=shib"
	loggedinURLPrefix = "https://studip.uni-passau.de/studip/dispatch.php"
)

// Verify the user is logged in
func verifyLoggedIn(resp *http.Response) bool {
	if resp.Request != nil && resp.Request.URL != nil {
		return strings.HasPrefix(resp.Request.URL.String(), loggedinURLPrefix)
	}
	return false
}

// Login performs a full SAML Client authentication to studip.
// If the login could not be completed, an error is returned.
func (a *API) Login(username, password string) error {

	// First request redirects either to studip (already logged in) or to SSO
	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return err
	}
	a.applyHeader(req)

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	if code := resp.StatusCode; code != http.StatusOK {
		return &StatusCodeError{
			Code: code,
			Msg:  fmt.Sprintf("Initial auth prepare request failed with status code: %d", code),
		}
	}

	// Check if already logged in
	if verifyLoggedIn(resp) {
		return nil
	}

	respLoginBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for login form
	if !reLoginForm.Match(respLoginBody) {
		return &APIError{
			Msg: "Could not find login form",
		}
	}

	// Login SSO

	// Next request url is last redirected url
	authurl := resp.Request.URL.String()

	authForm := url.Values{}
	authForm.Add("j_username", username)
	authForm.Add("j_password", password)

	reqAuth, err := http.NewRequest("POST", authurl, strings.NewReader(authForm.Encode()))
	if err != nil {
		return err
	}
	a.applyHeader(reqAuth)
	reqAuth.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respAuth, err := a.Client.Do(reqAuth)
	if err != nil {
		return err
	}
	if code := respAuth.StatusCode; code != http.StatusOK {
		return &StatusCodeError{
			Code: code,
			Msg:  fmt.Sprintf("Auth request failed with status code: %d", code),
		}
	}
	respAuthBody, err := ioutil.ReadAll(respAuth.Body)
	if err != nil {
		return err
	}
	defer respAuth.Body.Close()

	// Check for SAML Response page (SAML Confirmation form)
	// Otherwise username or password might be wrong
	m := reAuth.FindStringSubmatch(string(respAuthBody))

	// No login form
	if m == nil {
		invalidLoginMatch := reInvalidLogin.FindStringSubmatch(string(respAuthBody))
		if invalidLoginMatch != nil && len(invalidLoginMatch) == 2 {
			return &APIError{
				Msg:          fmt.Sprintf("Invalid login: %s", strings.TrimSpace(invalidLoginMatch[1])),
				InvalidLogin: true,
			}
		}
		return &APIError{
			Msg: "Could not finalize SAML Authentication, System down?",
		}
	}
	if len(m) != 6 {
		return &APIError{
			Msg: "Could not parse SAML Response form",
		}
	}

	samlRespURL := html.UnescapeString(m[1])
	field1Name := html.UnescapeString(m[2])
	field1Value := html.UnescapeString(m[3])
	field2Name := html.UnescapeString(m[4])
	field2Value := html.UnescapeString(m[5])

	//build form
	form := url.Values{}
	form.Add(field1Name, field1Value)
	form.Add(field2Name, field2Value)

	// Send SAML Response form, should redirect to studip
	reqSAMLResponse, err := http.NewRequest("POST", samlRespURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	a.applyHeader(reqSAMLResponse)
	reqSAMLResponse.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respSAMLResp, err := a.Client.Do(reqSAMLResponse)
	if err != nil {
		return err
	}
	if code := respSAMLResp.StatusCode; code != http.StatusOK {
		return &StatusCodeError{
			Code: code,
			Msg:  fmt.Sprintf("SAML Response request failed with status code: %d", code),
		}
	}
	if !(verifyLoggedIn(respSAMLResp)) {
		return &APIError{
			Msg: "Not redirected to studip after login",
		}
	}
	return nil
}
