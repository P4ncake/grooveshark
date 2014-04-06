# Grooveshark Public API Go Library

GsAPI offers Groovesharks Public API support in GO

## Getting Started

Install gsAPI package

~~~
go get github.com/P4ncake/gsAPI
~~~ 

~~~ go
package main 

import "github.com/P4ncake/gsAPI"

func main() {
	gsAPI.New("example", "1a79a4d60de6718e8e5b326e338ae533")
	sessionID, err := gsAPI.StartSession()
	if err != nil {
		panic(err)
	}

	_, err = gsAPI.Authenticate("user", "login")
	if err != nil {
		// Auth Failed
		panic(err)
	}
	/*
	 * Do something 
	 */

}

~~~
