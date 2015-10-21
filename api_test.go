package studip

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
)

// Adjust this values for tests
var (
	username = os.Getenv("STUDIP_USERNAME")
	password = os.Getenv("STUDIP_PASSWORD")
)

func TestAPILogin(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Could not create cookie jar: %s\n", err)
	}

	client := &http.Client{}
	client.Jar = jar
	api := &API{
		Client: client,
	}

	err = api.Login(username, password)
	if err != nil {
		t.Fatalf("Login failed: %s\n", err)
	}
}

func TestDocumentTree(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Could not create cookie jar: %s\n", err)
	}

	client := &http.Client{}
	client.Jar = jar
	api := &API{
		Client: client,
	}

	err = api.Login(username, password)
	if err != nil {
		t.Fatalf("Login failed: %s\n", err)
	}

	tree, err := api.DocumentTree()
	if err != nil {
		t.Fatalf("Error: %s\n", err)
	}
	if tree == nil || len(*tree) == 0 {
		t.Fatalf("Invalid or empty tree returned: %q", tree)
	}
}
