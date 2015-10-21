package studip

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
)

// Adjust this values for tests
var (
	username = os.Getenv("STUDIP_USERNAME")
	password = os.Getenv("STUDIP_PASSWORD")
	fileID   = os.Getenv("STUDIP_FILEID")
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

func TestGetFile(t *testing.T) {
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

	reader, err := api.GetFile(fileID)
	if err != nil {
		t.Fatalf("Error get file: %s\n", err)
	}
	defer reader.Close()

	f, err := ioutil.TempFile("", "studip")
	if err != nil {
		t.Fatalf("Error creating tmp file: %s\n", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	written, err := io.Copy(f, reader)
	if err != nil {
		t.Fatalf("Error writing contents to disk: %s\n", err)
	}
	if written < 128 {
		t.Fatalf("Invalid file content\n")
	}
}

func TestGetInvalidFile(t *testing.T) {
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

	_, err = api.GetFile("invalid")
	if err == nil {
		t.Fatalf("No error get invalid file")
	}
	if statuserr, ok := err.(*StatusCodeError); ok {
		if statuserr.Code != http.StatusNotFound {
			t.Fatalf("Error code should be http.StatusNotFound, got: %d\n", statuserr.Code)
		}
	} else {
		t.Fatalf("No StatusCodeError received: %s\n", err)
	}
}
