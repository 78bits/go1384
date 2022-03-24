package astm1384

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func Unmarshal2(messageData []byte, target interface{}, enc Encoding, tz Timezone, pv ProtocolVersion) error {

	//TODO: character conversion (copy from 1)

	// t := reflect.TypeOf(target).Elem()

	// currentLine := 0
	buffer := strings.Split(string(messageData), string([]byte{0x0D})) // copy
	// precautiously strip the 0A Linefeed
	for i := 0; i < len(buffer); i++ {
		buffer[i] = strings.Trim(buffer[i], string([]byte{0x0A}))
	}

	_, _, err := seqScan(buffer, 1 /*recursion-depth*/, 0, target, enc, tz, pv)
	if err != nil {
		return err
	}

	return nil
}

type RETV int

const OK RETV = 1
const UNEXPECTED RETV = 2
const ERROR RETV = 3

/* Scan Structure recursive. Note there are only 10b type of people: those that understand recursions, and those who dont */
func seqScan(buffer []string, depth int, currentLine int, target interface{}, enc Encoding, tz Timezone, pv ProtocolVersion) (int, RETV, error) {

	outerStructureType := reflect.TypeOf(target).Elem()

	fmt.Printf("seqScan (%s, %d)\n", reflect.TypeOf(target).Name(), depth)

	for i := 0; i < outerStructureType.NumField(); i++ {

		astmTag := outerStructureType.Field(i).Tag.Get("astm")
		astmTagsList := strings.Split(astmTag, ",")

		if len(astmTagsList) < 1 {
			continue // not annotated
		}

		// no tags provided means a neted array with more records or ignore
		if len(astmTagsList[0]) < 1 {

			// Not annotated array. If its a struct have to recurse, otherwise skip
			if outerStructureType.Field(i).Type.Kind() == reflect.Slice {

				// What is the type of the Slice? (Struct or string ?)
				sliceFieldType := reflect.TypeOf(outerStructureType.Field(i))

				// Array of Structs
				if sliceFieldType.Kind() == reflect.Struct {
					innerStructureType := outerStructureType.Field(i).Type.Elem()

					sliceForNestedStructure := reflect.MakeSlice(outerStructureType.Field(i).Type, 0, 0)

					for {

						allocatedElement := reflect.New(innerStructureType)
						var err error
						var retv RETV
						currentLine, retv, err = seqScan(buffer, depth+1, currentLine, allocatedElement.Interface(), enc, tz, pv)
						if err != nil {
							if retv == UNEXPECTED {
								if depth > 1 {
									// if nested structures abort due to unexpected records that does not creaate an error
									// as the parse will be conitnued one level higher
									return currentLine, UNEXPECTED, err
								} else {
									return currentLine, ERROR, err
								}
							}
						}

						sliceForNestedStructure = reflect.Append(sliceForNestedStructure, allocatedElement.Elem())
						reflect.ValueOf(target).Elem().Field(i).Set(sliceForNestedStructure)
					}
				}
			}
		}

		// A regular nested structure for scanning with a valid annotation
		/* if outerStructureType.Field(i).Type.Kind() == reflect.Struct {
			fmt.Println("DOWN")
			currentLine, retv, err := seqScan(buffer, depth+1, currentLine,
				reflect.ValueOf(target).Field(i).Elem(), enc, tz, pv)
			if err != nil {
				if retv == UNEXPECTED {
					if depth > 1 {
						// if nested structures abort due to unexpected records that does not creaate an error
						// as the parse will be conitnued one level higher
						return currentLine, UNEXPECTED, err
					} else {
						return currentLine, ERROR, err
					}
				}
			}
		}
		continue // empty annotation
		*/
		//TODO: Only valid if: H,L,M,P,O,S,C
		expectRecordType := astmTagsList[0][0]
		//h		fmt.Printf("I am expecting %c depth:%d \n", expectRecordType, depth)

		optional := false
		if contains(astmTagsList, "optional") {
			optional = true
		}

		// field := reflect.ValueOf(target).Elem().Field(i)
		// fieldValue := field.Interface()

		if expectRecordType == buffer[currentLine][0] {
			//TODO scan it
			//fmt.Printf("Consumed %c\n", expectRecordType)
			err := MapRecordFromString(expectRecordType, buffer[currentLine], outerStructureType.Field(i))
			if err != nil {
				return currentLine, ERROR, err
			}
			currentLine = currentLine + 1
		} else {
			if optional {
				fmt.Printf("Skipped optional %c\n", expectRecordType)
				continue
			} else {
				return currentLine, UNEXPECTED, errors.New(fmt.Sprintf("Expected Record-Type '%c' input was '%c' in depth (%d) (Abort)", expectRecordType, buffer[currentLine][0], depth))
			}
		}
		/*
			switch astmTagsList[0] {
			case "H":
				if err_inner := Scan(buffer[currentLine], fieldValue, 'H', optional); err_inner == nil {
					currentLine = currentLine + 1
				} else {
					return err_inner
				}
			case "M":
				if err_inner := Scan(buffer[currentLine], fieldValue, 'M', optional); err_inner == nil {
					currentLine = currentLine + 1
				} else {
					return err_inner
				}
			case "L":
				if err_inner := Scan(buffer[currentLine], fieldValue, 'L', optional); err_inner == nil {
					currentLine = currentLine + 1
				} else {
					return err_inner
				}
			default:
				// Unlabeled or wrong
				if field.Kind() == reflect.Struct {
					fmt.Println("Struct found !! (how about its an array ?) ", field.Type().Name())
				}
			}
		*/
		if currentLine >= len(buffer) {
			break
		}
	}

	return currentLine, OK, nil
}

func MapRecordFromString(recordtype byte, inputstr string, target interface{}) error {

	fmt.Printf("DECODE %s : %s\n", string(recordtype), inputstr)
	return nil
}

func Scan(input string, target interface{}, expectRecord byte, optional bool) error {
	if len(input) < 1 {
		return errors.New(fmt.Sprintf("Empty input. Excpected Record:%c", expectRecord))
	}
	if input[0] != expectRecord {
		return errors.New(fmt.Sprintf("Input Stream presentet Record type '%c' but expected '%c'", input[0], expectRecord))
	}
	fmt.Println("Scanning: ", input)
	return nil
}

func contains(list []string, search string) bool {
	for _, x := range list {
		if x == search {
			return true
		}
	}
	return false
}
