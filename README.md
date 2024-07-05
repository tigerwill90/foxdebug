[![Go Reference](https://pkg.go.dev/badge/github.com/tigerwill90/foxdebug.svg)](https://pkg.go.dev/github.com/tigerwill90/foxdebug)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tigerwill90/foxdebug)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tigerwill90/foxdebug)

# Convenient debug handler for fox.

Foxdebug is a small helper package for the [Fox](github.com/tigerwill90/fox) router, designed to provide detailed system and 
request information for debugging purposes. This package should only be used in a development environment as it may expose 
sensitive information.

**Installation**
````shell
go get -u github.com/tigerwill90/foxdebug
````

### Usage
To use the `foxdebug` package, simply import it and register the DebugHandler to any route.
````go
f := fox.New(fox.DefaultOptions())
f.MustHandle(http.MethodGet, "/debug", foxdebug.DebugHandler())

log.Fatal(http.ListenAndServe(":8080", f))
````

### License
This project is licensed under the MIT License - see the LICENSE file for details.
