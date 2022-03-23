package astm1384

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	charmap "github.com/aglyzov/charmap"
)

type ProtocolVersion int

const LIS2A2 ProtocolVersion = 2

type Encoding int

const Encoding_UTF8 Encoding = 1
const Encoding_ASCII Encoding = 2
const Encoding_Windows1250 Encoding = 3
const Encoding_Windows1251 Encoding = 4
const Encoding_Windows1252 Encoding = 5
const Encoding_DOS852 Encoding = 6
const Encoding_DOS855 Encoding = 7
const Encoding_DOS866 Encoding = 8

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

	location, err := time.LoadLocation(string(tz))
	if err != nil {
		return nil, err
	}

	var messagestr string
	switch enc {
	case Encoding_UTF8:
		// do nothing, this is correct
		messagestr = string(messageData)
	case Encoding_ASCII:
		messagestr = string(messageData)
	case Encoding_DOS866:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "DOS866")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	case Encoding_DOS855:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "DOS855")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	case Encoding_DOS852:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "DOS852")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	case Encoding_Windows1250:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "CP1250")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	case Encoding_Windows1251:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "CP1251")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	case Encoding_Windows1252:
		messagebytes, err := charmap.ANY_to_UTF8(messageData, "CP1252")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid input : %s", err))
		}
		messagestr = string(messagebytes)
	default:
		return nil, errors.New(fmt.Sprintf("Invalid Codepage %d", enc))
	}

	tokeninput, err2 := astm1384Scanner(messagestr, location, "\n")
	if err2 != nil {
		return nil, err2
	}

	message, err := parseAST(tokeninput)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func Marshal(message *ASTMMessage, enc Encoding, tz Timezone, pv ProtocolVersion) ([]byte, error) {
	var buffer bytes.Buffer

	location, err := time.LoadLocation(string(tz))
	if err != nil {
		return []byte{}, err
	}

	err = convertToASTMFileRecord("H", message.Header, []string{"\\", "^", "&"}, location, &buffer)
	if err != nil {
		log.Println(err)
		return []byte{}, errors.New(fmt.Sprintf("Failed to marshal header: %s", err))
	}
	buffer.Write([]byte{10, 13})

	if message.Manufacturer != nil {
		err := convertToASTMFileRecord("M", message.Manufacturer, []string{"\\", "^", "&"}, location, &buffer)
		if err != nil {
			log.Println(err)
			return []byte{}, errors.New(fmt.Sprintf("Failed to marshal manufacturer-record: %s", err))
		}
		buffer.Write([]byte{10, 13})
	}

	for i, record := range message.Records {
		if record.Patient != nil {
			record.Patient.SequenceNumber = i + 1
			err := convertToASTMFileRecord("P", record.Patient, []string{"|", "^", "&"}, location, &buffer)
			if err != nil {
				log.Println(err)
				return []byte{}, errors.New(fmt.Sprintf("Failed to marshal header: %s", err))
			}
			buffer.Write([]byte{10, 13})

			for orderresults_i, orderresults := range record.OrdersAndResults {
				orderresults.Order.SequenceNumber = orderresults_i + 1
				err := convertToASTMFileRecord("O", orderresults.Order, []string{"|", "^", "&"}, location, &buffer)
				if err != nil {
					log.Println(err)
					return []byte{}, errors.New(fmt.Sprintf("Failed to marshal order-records: %s", err))
				}
				buffer.Write([]byte{10, 13})

				for results_i, result := range orderresults.Results {
					result.Result.SequenceNumber = results_i + 1
					err := convertToASTMFileRecord("R", result.Result, []string{"|", "^", "&"}, location, &buffer)
					if err != nil {
						log.Println(err)
						return []byte{}, errors.New(fmt.Sprintf("Failed to marshal result-record %s", err))
					}
					buffer.Write([]byte{10, 13})
					for comment_i, comment := range result.Comments {
						comment.SequenceNumber = comment_i + 1
						err := convertToASTMFileRecord("C", comment, []string{"|", "^", "&"}, location, &buffer)
						if err != nil {
							log.Println(err)
							return []byte{}, errors.New(fmt.Sprintf("Failed to marshal result-comment %s", err))
						}
						buffer.Write([]byte{10, 13})
					}
				}
			}
		}
	}
	buffer.Write([]byte("L|1|N"))
	buffer.Write([]byte{10, 13})

	return buffer.Bytes(), nil
}

func convertToASTMFileRecord(recordtype string, target interface{}, delimiter []string, tz *time.Location, buffer *bytes.Buffer) error {

	t := reflect.TypeOf(target).Elem()

	entries := make(map[int]string, 0)

	maxIdx := 0

	for i := 0; i < t.NumField(); i++ {
		astmTag := t.Field(i).Tag.Get("astm")
		astmTagsList := strings.Split(astmTag, ",")
		if len(astmTagsList) == 0 || astmTag == "" {
			continue // nothing to process when someone requires astm:
		}
		idx, err := strconv.Atoi(astmTagsList[0])
		idx = idx - 1
		if idx < 0 {
			return errors.New(fmt.Sprintf("Illegal annotation <%s> in for field %s", astmTag, t.Name()))
		}
		if err != nil {
			return err
		}
		if idx > maxIdx {
			maxIdx = idx
		}

		isLongDate := false
		for i := 0; i < len(astmTagsList); i++ {
			if astmTagsList[i] == "longdate" {
				isLongDate = true
			}
		}

		field := reflect.ValueOf(target).Elem().Field(i)
		fieldValue := field.Interface()

		switch fieldValue.(type) {
		case int:
			entries[idx] = strconv.Itoa(int(field.Int()))
		case string:
			entries[idx] = string(field.String())
		case []string:
			arry := fieldValue.([]string)
			outstring := ""
			for i := 0; i < len(arry); i++ {
				outstring = outstring + arry[i]
				if i < len(arry)-1 {
					outstring = outstring + "^"
				}
			}
			entries[idx] = outstring
		case [][]string:
			arrym := fieldValue.([][]string)
			outstring := ""
			for i := 0; i < len(arrym); i++ {
				subarray := arrym[i]
				for j := 0; j < len(subarray); j++ {
					outstring = outstring + subarray[j]
					if j < len(subarray)-1 {
						outstring = outstring + "^"
					}
				}
				if i < len(arrym)-1 {
					outstring = outstring + "\\"
				}
			}
			entries[idx] = outstring
		case time.Time:
			if fieldValue.(time.Time).IsZero() {
				// dates can be zero = no output
				break
			}
			if isLongDate {
				entries[idx] = fieldValue.(time.Time).In(tz).Format("20060102150405")
			} else {
				entries[idx] = fieldValue.(time.Time).Format("20060102")
			}
		default:
			return errors.New(fmt.Sprintf("Unsupported field type %s", field.Type()))
		}
	}

	output := recordtype + "|"
	for i := 0; i <= maxIdx; i++ {
		value := entries[i]
		output = output + value
		if i < maxIdx {
			output = output + "|"
		}
	}
	buffer.Write([]byte(output))
	return nil
}
