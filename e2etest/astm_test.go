package main

import (
	"astm1394/astm1384"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"log"

	"github.com/stretchr/testify/assert"
)

func TestReadfileBeCom52(t *testing.T) {
	fileData, err := ioutil.ReadFile("../protocoltest/becom/5.2/bloodtype.astm")
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

	assert.Equal(t, "IH v5.2", message.Header.SenderStreetAddress)
	assert.Equal(t, "Bio-Rad", message.Header.SenderNameOrID)
	locale, err := time.LoadLocation("Europe/Berlin")
	assert.Nil(t, err)
	localtime := message.Header.DateAndTime.In(locale)
	assert.Equal(t, "20220315194227", localtime.Format("20060102150405"))

	assert.Equal(t, 1 /*Patient*/, len(message.Records))
	assert.Equal(t, "1010868845", message.Records[0].Patient.LabAssignedPatientID)
	assert.Equal(t, "Testus", message.Records[0].Patient.LastName)
	assert.Equal(t, "19400607", message.Records[0].Patient.DOB.Format("20060102"))
	assert.Equal(t, "M", message.Records[0].Patient.Gender)
	assert.Equal(t, "Test", message.Records[0].Patient.FirstName)
	assert.Equal(t, "MO10", message.Records[0].Orders[0].Order.UniversalTestID_ManufacturerCode)

	/* fmt.Printf("Messageheader: %+v\n", message.Header)
	   for _, record := range message.Records {
		fmt.Printf("Patient : %s, %s\n", record.Patient.Name[0], record.Patient.Name[1])
		for _, order := range record.Orders {
			fmt.Printf("  Order: %+v\n", order.Order)
			for _, result := range order.Results {
				fmt.Printf("   Result: %+v\n", result.Result)
			}
		}
	}*/
}

func noTestReadfileEuroImmunAnalyzer1(t *testing.T) {
	fileData, err := ioutil.ReadFile("../protocoltest/euroimmun/sampleigg.astm")
	if err != nil {
		log.Print(err)
		t.Fail()
		return
	}

	message, err := astm1384.Unmarshal(fileData,
		astm1384.Encoding_Windows1252, astm1384.Timezone_EuropeBerlin, astm1384.LIS2A2)
	if err != nil {
		log.Print(err)
		t.Fail()
		return
	}

	assert.Equal(t, len(message.Records), 20)
	assert.Equal(t, 1, message.Records[0].Patient.SequenceNumber)
	assert.Equal(t, "TEST-27-079-5-1", message.Records[0].Patient.LabAssignedPatientID)
	assert.Equal(t, 1, message.Records[0].Orders[0].Order.SequenceNumber)
	assert.Equal(t, "SARSCOV2IGA", message.Records[0].Orders[0].Order.UniversalTestID[4])
	assert.Equal(t, "SARSCOV2IGA", message.Records[0].Orders[0].Order.UniversalTestID_ManufacturerCode)

	// The manufacturer has a lot# in his Additional Fields
	assert.Equal(t, "28343", message.Records[0].Orders[0].Order.UniversalTestID_Custom2)

	testTimeAsInFile, err := time.Parse(time.RFC3339, "2022-02-18T09:07:37+01:00")
	assert.Nil(t, err)
	timezone, err := time.LoadLocation("Europe/Berlin")
	assert.Nil(t, err)
	testTimeAsInFile = testTimeAsInFile.In(timezone)
	assert.Equal(t, testTimeAsInFile, message.Records[0].Orders[0].Order.RequestedOrderDateTime)

}
