package astm1384

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type TokenType int

const TokenRoot TokenType = 0
const TokenHeader TokenType = 1
const TokenPatient TokenType = 2
const TokenComment TokenType = 3
const TokenOrder TokenType = 4
const TokenResult TokenType = 5
const TokenTerminator TokenType = 6
const TokenManufacturer TokenType = 7

type Token struct {
	Type TokenType
	Data interface{}
	Next *Token
}

func astm1384Scanner(messagestr string,
	timezone *time.Location,
	defaultLinebreak string) (*Token, error) {

	delimiters := "|" // start with default-delimiter

	root := &Token{Type: TokenRoot, Next: nil}
	tail := root

	for _, line := range strings.Split(messagestr, defaultLinebreak) {

		fields := strings.Split(line, "|")
		if len(fields) < 1 {
			continue
		}

		for i := 0; i < len(fields); i++ {
			fields[i] = strings.Trim(fields[i], "\r\n\t")
		}

		switch fields[0] {
		case "H":
			var h Header
			err := reflect_map(fields, &h, "|", timezone)
			if err != nil {
				return nil, err
			}
			delimiters = h.Delimiters

			token := &Token{Type: TokenHeader, Data: &h}
			tail.Next = token
			tail = token
		case "P":
			var p Patient
			err := reflect_map(fields, &p, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenPatient, Data: &p}
			tail.Next = token
			tail = token
		case "O":
			var o Order
			err := reflect_map(fields, &o, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenOrder, Data: &o}
			tail.Next = token
			tail = token
		case "M":
			var m Manufacturer
			err := reflect_map(fields, &m, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenManufacturer, Data: &m}
			tail.Next = token
			tail = token
		case "C":
			var c Comment
			err := reflect_map(fields, &c, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenComment, Data: &c}
			tail.Next = token
			tail = token
		case "R":
			var r Result
			err := reflect_map(fields, &r, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenResult, Data: &r}
			tail.Next = token
			tail = token
		case "L":
			var t Terminator
			err := reflect_map(fields, &t, delimiters, timezone)
			if err != nil {
				return nil, err
			}
			token := &Token{Type: TokenTerminator, Data: &t}
			tail.Next = token
			tail = token
		default:
			return nil, errors.New(fmt.Sprintf("Invalid Record Identifier : '%s'", fields[0]))
		}
	}

	return root, nil
}

func reflect_map(fields []string, target interface{},
	usedelimiter string, timezone *time.Location) error {

	t := reflect.TypeOf(target).Elem()

	for i := 0; i < t.NumField(); i++ {
		astmTag := t.Field(i).Tag.Get("astm")

		astmTagsList := strings.Split(astmTag, ",")

		if len(astmTagsList) == 0 || astmTag == "" {
			continue // nothing to process when someone requires astm:
		}

		mapFieldNo, err := strconv.Atoi(astmTagsList[0]) // just a number
		if err != nil {
			return err
		}

		if mapFieldNo >= len(fields) {
			continue // mapped field is beyond the data
		}

		field := reflect.ValueOf(target).Elem().Field(i)
		fieldValue := field.Interface()

		switch fieldValue.(type) {
		case string:
			if len(astmTagsList) > 1 {
				// further subdivide like this part "|^^^MO10^^28343^|"
				subFields := strings.Split(fields[mapFieldNo], "^")
				subFieldNo, err := strconv.Atoi(astmTagsList[1])
				if err != nil {
					return errors.New(fmt.Sprintf("Invalid annotation astm:%s. %s", astmTag, err))
				}
				if len(subFields) >= subFieldNo && subFieldNo >= 1 {
					field.SetString(subFields[subFieldNo-1])
				} else {
					// when fields are not present they just dont get mapped = skipping here
				}
			} else {
				field.SetString(fields[mapFieldNo])
			}
		case int:
			num, err := strconv.Atoi(fields[mapFieldNo])
			if err != nil {
				return err
			}
			field.SetInt(int64(num))
		case []string:
			instr := fields[mapFieldNo]
			list := splitany(instr, usedelimiter)
			field.Set(reflect.ValueOf(list))
		case time.Time:
			instr := fields[mapFieldNo]
			if instr == "" {
				field.Set(reflect.ValueOf(time.Time{}))
			} else if len(instr) == 8 { // YYYYMMDD See Section 5.6.2 https://samson-rus.com/wp-content/files/LIS2-A2.pdf
				time_located, err := time.ParseInLocation("20060102", instr, timezone)
				if err != nil {
					return errors.New(fmt.Sprintf("Invalid time format <%s>", instr))
				}
				field.Set(reflect.ValueOf(time_located))
			} else if len(instr) == 14 { // YYYYMMDDHHMMSS
				time_located, err := time.ParseInLocation("20060102150405", instr, timezone)
				if err != nil {
					return errors.New(fmt.Sprintf("Invalid time format <%s>", instr))
				}
				field.Set(reflect.ValueOf(time_located.UTC()))
			} else {
				return errors.New(fmt.Sprintf("Unrecognized time format <%s>", instr))
			}
		default:
			return errors.New(fmt.Sprintf("Invalid field-Type '%s' for mapping (not implemented)", t.Field(i).Type))
		}
	}

	return nil
}

/** splitany - split string by any of the delimiters
**/
func splitany(s string, delimiters string) []string {

	ret := []string{}

	for {
		pos := strings.IndexAny(s, delimiters)
		if pos < 0 {
			break
		}
		cut := s[0:pos]
		s = s[pos+1:]
		ret = append(ret, cut)
	}

	if len(s) > 0 {
		ret = append(ret, s)
	}

	return ret
}
