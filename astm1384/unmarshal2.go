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

	_, err := seqScan(buffer, 0, target, enc, tz, pv)
	if err != nil {
		return err
	}

	return nil
}

/* Scan Structure recursive. Note there are only 10b type of people: those that understand recursions, and those who dont */
func seqScan(buffer []string, currentLine int, target interface{}, enc Encoding, tz Timezone, pv ProtocolVersion) (int, error) {

	outerStructure := reflect.TypeOf(target).Elem()

	for i := 0; i < outerStructure.NumField(); i++ {

		astmTag := outerStructure.Field(i).Tag.Get("astm")
		astmTagsList := strings.Split(astmTag, ",")

		if len(astmTagsList) < 1 {
			continue // not annotated
		}
		if len(astmTagsList[0]) < 1 {
			// Not annotated array. If its a struct have to recurse, otherwise skip
			if outerStructure.Field(i).Type.Kind() == reflect.Slice {
				sliceOfStructs := reflect.TypeOf(outerStructure.Field(i)) // this is the slice
				innerTypeOfStruct := outerStructure.Field(i).Type.Elem()

				slice := reflect.MakeSlice(outerStructure.Field(i).Type, 0, 0)
				fmt.Println("Da Slice : ", slice)

				if sliceOfStructs.Kind() == reflect.Struct {

					for {
						fmt.Println("Loop")

						allocatedElement := reflect.New(innerTypeOfStruct)
						var err error
						currentLine, err = seqScan(buffer, currentLine, allocatedElement.Interface(), enc, tz, pv)
						if err != nil {
							fmt.Println("Exit due to failed read")
							// return currentLine, err
							break
						}

						// append to array
						// slice := reflect.ValueOf(target).Elem().Field(i)
						// fmt.Printf("Outer %s\n", slice.Kind())
						// os.Exit(-1)
						// reflect.ValueOf(outerType.Field(i)).Set(reflect.Append(reflect.ValueOf(outerType.Field(i)), allocatedElement))
						// reflect.ValueOf(slice).Set()

						slice = reflect.Append(slice, allocatedElement.Elem())
						// fmt.Printf("Appended %s\n", appended)
						// aslice := reflect.ValueOf(slice)
						//fmt.Printf("QUESTIONMARK %s\n", appended.Interface())
						// reflect.ValueOf()

						//fmt.Println("THIS TIME DONT OVERLOOK THISE ONE")
						//reflect.ValueOf(slice).Set(reflect.ValueOf(slice.Elem()))
						//os.Exit(-1)
						// reflect.ValueOf(slice.Interface()).Set(
						//	reflect.Append(reflect.ValueOf(slice.Interface()), allocatedElement.Elem()))
						// slice.Set(reflect.Append(reflect.ValueOf(slice.Interface()), allocatedElement.Elem()))
						// reflect.ValueOf(slice).Set(reflect.Append(reflect.ValueOf(slice.Interface()), allocatedElement.Elem()))
					}

					//reflect.ValueOf(outerStructure.Field(i)).Set(slice)
					reflect.ValueOf(outerStructure.Field(i)).
				}
			}
			continue // empty annotation
		}

		//TODO: Only valid if: H,L,M,P,O,S,C
		expectRecordType := astmTagsList[0][0]
		fmt.Printf("I am expecting %c\n", expectRecordType)

		optional := false
		if contains(astmTagsList, "optional") {
			optional = true
		}

		// field := reflect.ValueOf(target).Elem().Field(i)
		// fieldValue := field.Interface()

		if expectRecordType == buffer[currentLine][0] {
			//TODO scan it
			fmt.Printf("Consumed %c\n", expectRecordType)
			currentLine = currentLine + 1
		} else {
			if optional {
				fmt.Printf("Skipped optional %c\n", expectRecordType)
				continue
			} else {
				fmt.Println("Failed a bit ...")
				return currentLine, errors.New(fmt.Sprintf("Expected Record-Type '%c' input was '%c' (Abort)", expectRecordType, buffer[currentLine][0]))
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

	return currentLine, nil
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
