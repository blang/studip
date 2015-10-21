studip library for golang 
======

studip is a library to communicate with the [StudIP](http://www.studip.de/) platform implemented by the Univerity Passau. It works with Shibboleth SAML authentication and contains adjustments to work with this specific site. It could be adapted to other StudIP platforms with similar login mechanisms.

Usage
-----
```bash
$ go get github.com/blang/studip
```
Note: Always vendor your dependencies or fix on a specific version tag.

```go
import github.com/blang/studip

jar, _ := cookiejar.New(nil)
client := &http.Client{}
client.Jar = jar

api := &studip.API{
    Client: client,
}

err = api.Login(username, password)
if err != nil {
    t.Fatalf("Login failed: %s\n", err)
}

tree, _ := api.DocumentTree()
```

Also check the [GoDocs](http://godoc.org/github.com/blang/studip).

Features
-----

- Shibboleth SAML Login
- DocumentTree

Motivation
-----

A working studip api implementation could be used to create several services like automatic document download or notifications about new entries.
I simply couldn't find any lib supporting the studip api. 

Contribution
-----

Feel free to make a pull request. For bigger changes create a issue first to discuss about it.


License
-----

See [LICENSE](LICENSE) file.
