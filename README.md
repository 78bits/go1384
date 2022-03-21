# go1384 ![build status](https://travis-ci.org/78bit/uuid.svg?branch=master)

Golang library for handling ASTM lis2a2 Procotol

###### Install
`go get github.com/78bits/astm1384`

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