# go1384 ![build status](https://travis-ci.org/78bit/uuid.svg?branch=master)
The uuid package generates and inspects UUIDs based on
[RFC 4122](http://tools.ietf.org/html/rfc4122)
and DCE 1.1: Authentication and Security Services. 

This package is based on the github.com/pborman/uuid package (previously named
code.google.com/p/go-uuid).  It differs from these earlier packages in that
a UUID is a 16 byte array rather than a byte slice.  One loss due to this
change is the ability to represent an invalid UUID (vs a NIL UUID).

###### Install
`go get github.com/78bits/astm1384`

# go1384
Golang library for handling ASTM Procotol 

## Features

  - Encoding Support (UTF-8/Codepage 437/Codepage 850/Windows 1254) support
  - Timezone Support
  - Marshal/Unmarshal function

## Installation

Install the package with the following command.

``` shell
go get github.com/78bits/go1384/...
```
## Quick Start

The following Go code decodes a ASTM-File.

``` go
    fileData, err := ioutil.ReadFile("protocoltest/becom/5.2/bloodtype.astm")
	if err != nil {
		log.Fatal(err)		
	}

	message, err := astm1384.Unmarshal(fileData,
		astm1384.Encoding_Windows1252, 
        astm1384.Timezone_EuropeBerlin, 
        astm1384.LIS2A2)
	if err != nil {
		log.Fatal(err)		
	}
```