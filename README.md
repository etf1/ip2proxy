# ip2proxy

IP2Location.com Proxy database parser for Golang

## Install & prerequisites

* Get the library:
```sh
go get -u github.com/etf1/ip2proxy
```
* Pick the database you need : https://www.ip2location.com/proxy-database
(You can try with a free demo database (lite) : https://lite.ip2location.com/)

* Save it unzipped to a directory readable by your Golang project

## Use it

```go
package main

import (
	"github.com/etf1/ip2proxy"
)

// let's go !
func main() {
	db, err := ip2proxy.Open("/where/you/unzipped/IP2PROXY-LITE-PX4.BIN")
	if err != nil {
		panic(err)
	}
	res, err := db.LookupIPV4Dot("2.7.154.188")
	if err != nil {
		panic(err)
	}
	if res.Proxy == ip2proxy.ProxyTOR {
		println("2.7.154.188 is a TOR output node !")
	} else {
	    println("2.7.154.188 is NOT a TOR output node")	
	}
}
```

