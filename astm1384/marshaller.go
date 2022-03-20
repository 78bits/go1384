package astm1384

import (
	"errors"
	"fmt"
	"time"
)

type ProtocolVersion int

const LIS2A2 ProtocolVersion = 2

type Encoding int

const Encoding_UTF8 Encoding = 1
const Encoding_ASCII Encoding = 2
const Encoding_Windows1252 Encoding = 3
const Encoding_DOS850 Encoding = 4
const Encoding_DOS437 Encoding = 5

type Timezone string

const Timezone_UTC Timezone = "UTC"
const Timezone_EuropeBerlin Timezone = "Europe/Berlin"
const Timezone_EuropeBudapest Timezone = "Europe/Budapest"
const Timezone_EuropeLondon Timezone = "Europe/London"

func Unmarshal(messageData []byte, enc Encoding, tz Timezone, pv ProtocolVersion) (*ASTMMessage, error) {

	switch pv {
	case ProtocolVersion(LIS2A2):
	default:
		return nil, errors.New("Protocol Not implemented")
	}

	timezone, err := time.LoadLocation(string(tz))
	if err != nil {
		return nil, err
	}

	switch enc {
	case Encoding_UTF8:
		// do nothing, this is correct
	case Encoding_ASCII:
		//TODO
	case Encoding_Windows1252:
		//TODO
	case Encoding_DOS850:
	case Encoding_DOS437:
	default:
		return nil, errors.New(fmt.Sprintf("Invalid Codepage %d", enc))
	}
	//TODO: finsih charset conversion
	//rInUTF8 := transform.NewReader(strings.NewReader(string(fileData)), charmap.Windows1252.NewDecoder())

	messagestr := string(messageData)

	tokeninput, err2 := astm1384Scanner(messagestr, timezone, "\n")
	if err2 != nil {
		return nil, err2
	}

	message, err := parseAST(tokeninput)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func Marshal(message *ASTMMessage) (string, error) {
	ret := ""
	/* 	for _, record := range message.GetRecords() {
	   		ret = ret + record.AsString()
	   	}
	*/return ret, nil
}
