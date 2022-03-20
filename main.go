package main

import (
	"astm1394/astm1384"
	"fmt"
	"io/ioutil"
)

func main() {

	fileData, err := ioutil.ReadFile("protocoltest/becom/5.2/bloodtype.astm")
	if err != nil {
		fmt.Println("Failed : ", err)
		return
	}

	message, err := astm1384.Unmarshal(fileData,
		astm1384.Encoding_Windows1252, astm1384.Timezone_EuropeBerlin, astm1384.LIS2A2)
	if err != nil {
		fmt.Println("Error in unmarshaling the message ", err)
		return
	}

	fmt.Printf("Messageheader: %+v\n", message.Header)
	for _, record := range message.Records {
		fmt.Printf("Patient : %s, %s\n", record.Patient.Name[0], record.Patient.Name[1])
		for _, order := range record.Orders {
			fmt.Printf("  Order: %+v\n", order.Order)
			for _, result := range order.Results {
				fmt.Printf("   Result: %+v\n", result.Result)
			}
		}
	}

	theoriginalMessage, err := astm1384.Marshal(message)
	fmt.Println(theoriginalMessage)
}
