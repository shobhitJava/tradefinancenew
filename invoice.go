package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// Invoice implements the document smart contract
type Invoice struct {
	PAYER                string
	PAYEE                string
	TAX_REGISTRY_NO      int
	INVOICE_CODE         int
	INVOICE_NUMBER       int
	PRINTING_NO          int
	Rows                 []invoiceRow
	TOTAL_IN_WORDS       string
	TOTAL_IN_FIGURES     int
	PRINT_NO             int
	ANTI_FORGERY_CODE    string
	DATE_ISSUED          string
	DUE_DATE             string
	SHIPPING_DATE        string
	LC_NUMBER            string
	DATE_OF_PRESENTATION string
	CURRENCY             string
}

//RowInvoice ...
type invoiceRow struct {
	//ID             string //`json:"id" bson:"id"`
	SERVICE        string
	ITEM           int
	AMOUNT_CHARGED int
	REMARKS        string
}

//Init initializes the document smart contract
func (t *Invoice) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// Check if table already exists
	_, err := stub.GetTable("invoiceTable")
	if err == nil {
		// Table already exists; do not recreate
		return nil, nil
	}

	// Create L/C Table
	err = stub.CreateTable("invoiceTable", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Type", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "UID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "DocJSON", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "DocPDF", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "Status", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating invoiceTable.")
	}

	return nil, nil

}

// isEarlierDate returns true if date1 is earlier than date2, false otherwise
// Assumes that date is presented in 'mm/dd/yyyy' format
func (t *Invoice) isEarlierDate(date1Str string, date2Str string) (bool, error) {
	layout := time_format

	// Parse the dates
	date1, err := time.Parse(layout, date1Str)
	if err != nil {
		return true, errors.New("Incorrect date format for date1. Expecting mm/dd/yyyy; " + date1Str)
	}
	date2, err := time.Parse(layout, date2Str)
	if err != nil {
		return true, errors.New("Incorrect date format for date2. Expecting mm/dd/yyyy; " + date2Str)
	}

	return date1.Before(date2) || date1.Equal(date2), nil
}

//ValidateDoc () – validates that the document is correct
func (t *Invoice) ValidateDoc(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2.")
	}

	// Prepare the response struct
	resultMap := make(map[string]string)

	// Specify the time format
	layout := time_format

	docJSON := []byte(args[0])
	lcJSON := []byte(args[1])

	var invoiceDataStruct Invoice
	err := json.Unmarshal(docJSON, &invoiceDataStruct)
	if err != nil {
		return nil, err
	}

	//Ensure that all fields are present
	if invoiceDataStruct.INVOICE_CODE < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.INVOICE_NUMBER < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.PRINTING_NO < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.PRINT_NO < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.TAX_REGISTRY_NO < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.TOTAL_IN_FIGURES < 0 {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.ANTI_FORGERY_CODE == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.CURRENCY == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.DATE_ISSUED == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.DATE_OF_PRESENTATION == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.DUE_DATE == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.LC_NUMBER == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.PAYEE == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.PAYER == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.SHIPPING_DATE == "" {
		return []byte("Error: Required field not provided."), nil
	} else if invoiceDataStruct.TOTAL_IN_WORDS == "" {
		return []byte("Error: Required field not provided."), nil
	} else if len(invoiceDataStruct.Rows) == 0 {
		return []byte("Error: Required field not provided."), nil
	}

	var lcStruct LC
	err = json.Unmarshal(lcJSON, &lcStruct)
	if err != nil {
		return nil, err
	}

	// Validation #1: Ensure LC number in LC and invoiceData match
	if invoiceDataStruct.LC_NUMBER != lcStruct.Tag20 {
		resultMap["result"] = "Error: LC number in invoice does not match the number on LC"
		fmt.Printf("Validation #1 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #1 passed\n")

	// Validation #2: Issue date of supporting document should not be earlier than issue date of LC
	check, err := t.isEarlierDate(invoiceDataStruct.DATE_OF_PRESENTATION, lcStruct.Tag31C)
	if err != nil {
		resultMap["result"] = "Error: " + err.Error()
		fmt.Printf("Validation #2 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	if check == true {
		resultMap["result"] = "Error: Date of presentation cannot be earlier than LC issue date"
		fmt.Printf("Validation #2 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}

	check, err = t.isEarlierDate(invoiceDataStruct.DATE_ISSUED, lcStruct.Tag31C)
	if err != nil {
		resultMap["result"] = "Error: " + err.Error()
		fmt.Printf("Validation #2 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	if check == true {
		resultMap["result"] = "Error: Invoice date cannot be earlier than LC issue date"
		fmt.Printf("Validation #2 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #2 passed\n")

	// Validation #3: Check that the total amount on the invoice is within the tolerance limit of the value of the L/C
	// Tag39A is tolerance
	toleranceValueStr := strings.Split(lcStruct.Tag39A, "/")[0]
	toleranceValue, err := strconv.Atoi(toleranceValueStr)
	if err != nil || toleranceValue < 0 || toleranceValue > 100 {
		resultMap["result"] = "Error: Tolerance value provided in L/C is not an integer between 0 and 100"
		fmt.Printf("Validation #3 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}

	// Tag32B is currency and amount. Extract amount first
	re := regexp.MustCompile("[0-9]+")
	amountStr := re.FindAllString(lcStruct.Tag32B, 1)[0]
	amount, _ := strconv.Atoi(amountStr)
	lowerLimit := (1 - float64(toleranceValue)/100) * float64(amount)
	upperLimit := (1 + float64(toleranceValue)/100) * float64(amount)

	invoiceAmount := invoiceDataStruct.TOTAL_IN_FIGURES

	if float64(invoiceAmount) < lowerLimit || float64(invoiceAmount) > upperLimit {
		resultMap["result"] = "Error: Total amount in invoice is not within tolerance limit specified in L/C"
		fmt.Printf("Validation #3 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #3 passed\n")

	// Validation #4,5: InvoiceDate in Invoice + Period of presentation in L/C <= DueDate in Invoice
	// Get period of presentation from Tag48 of L/C
	periodOfPresentationStr := re.FindAllString(lcStruct.Tag48, 1)[0]
	periodOfPresentation, _ := strconv.Atoi(periodOfPresentationStr)

	// Parse the dates
	invoiceDate, err := time.Parse(layout, invoiceDataStruct.DATE_ISSUED)
	if err != nil {
		resultMap["result"] = "Error: Incorrect date format for invoice date. Expecting mm/dd/yyyy"
		fmt.Printf("Validation #4 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}

	dueDate, err := time.Parse(layout, invoiceDataStruct.DUE_DATE)
	if err != nil {
		resultMap["result"] = "Error: Incorrect date format for invoice date. Expecting mm/dd/yyyy"
		fmt.Printf("Validation #4 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}

	// Add 0 years, 0 months, periodOfPresentation days to invoice date to obtain the calculatedDueDate
	calculatedDueDate := invoiceDate.AddDate(0, 0, periodOfPresentation)

	if dueDate.Before(calculatedDueDate) {
		resultMap["result"] = "Error: Invoice date + period of presentation as per L/C cannot be later than due date in invoice"
		fmt.Printf("Validation #4 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #4 passed\n")
	fmt.Printf("Validation #5 passed\n")

	// Validation #6: Shipping date should be no later than the latest date of shipment in L/C
	check, err = t.isEarlierDate(invoiceDataStruct.SHIPPING_DATE, lcStruct.Tag44C)
	if err != nil {
		resultMap["result"] = "Error: " + err.Error()
		fmt.Printf("Validation #6 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	if check == false {
		resultMap["result"] = "Error: Shipping date cannot be later than the latest date of shipment as per L/C"
		fmt.Printf("Validation #6 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #6 passed\n")

	// Validation #7: Presentation date should be earlier than shipping date + 21 days
	// Parse the shipping date
	shippingDate, err := time.Parse(layout, invoiceDataStruct.SHIPPING_DATE)
	if err != nil {
		resultMap["result"] = "Error: Incorrect date format for shipping date in invoice data. Expecting mm/dd/yyyy"
		fmt.Printf("Validation #7 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}

	// Add 0 years, 0 months, 21 days to shipping date to obtain the latestPresentationDate
	latestPresentationDate := shippingDate.AddDate(0, 0, 21)

	check, err = t.isEarlierDate(invoiceDataStruct.DATE_OF_PRESENTATION, latestPresentationDate.Format(layout))
	if err != nil {
		resultMap["result"] = "Error: " + err.Error()
		fmt.Printf("Validation #7 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	if check == false {
		resultMap["result"] = "Error: Presentation date cannot be later than shipping date + 21 days"
		fmt.Printf("Validation #7 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #7 passed\n")

	// Validation #8: Invoice date should not be earlier than L/C issuance date
	// Invoice date already checked as part of validation rule #2
	fmt.Printf("Validation #8 passed\n")

	// Validation #9: Currency in invoice data should match currency in L/C
	// Tag32B is currency and amount. Extract currency
	re = regexp.MustCompile("[A-Z]+")
	currency := re.FindAllString(lcStruct.Tag32B, 1)[0]

	if currency != invoiceDataStruct.CURRENCY {
		resultMap["result"] = "Error: Currency in invoice data does not match currency in L/C"
		fmt.Printf("Validation #9 failed\n")

		// Return the result as a JSON string
		return json.Marshal(resultMap)
	}
	fmt.Printf("Validation #9 passed\n")

	// All the validation checks have passed
	resultMap["result"] = "Success: All validation checks passed"

	// Return the result as a JSON string
	return json.Marshal(resultMap)

}

//SubmitDoc () – Calls ValidateDoc internally and upon success inserts a new row in the table
func (t *Invoice) SubmitDoc(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3.")
	}

	UID := args[0]
	docJSON := []byte(args[1])
	docPDF := []byte(args[2])

	//TODO call ValidateDoc instead
	//Make sure that args[1] is a JSON object
	var js map[string]interface{}
	err := json.Unmarshal(docJSON, &js)
	if err != nil {
		return nil, err
	}

	// Insert a row
	ok, err := stub.InsertRow("invoiceTable", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: "DOC"}},
			&shim.Column{Value: &shim.Column_String_{String_: UID}},
			&shim.Column{Value: &shim.Column_Bytes{Bytes: docJSON}},
			&shim.Column{Value: &shim.Column_Bytes{Bytes: docPDF}},
			&shim.Column{Value: &shim.Column_String_{String_: "SUBMITTED_BY_EB"}}},
	})

	if !ok && err == nil {
		return nil, errors.New("Document already exists.")
	}

	if err != nil {
		return nil, err
	}

	return nil, nil

}

//UpdateStatus () – Updates current document Status. Enforces Status transition logic.
func (t *Invoice) UpdateStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2.")
	}

	UID := args[0]
	newStatus := args[1]

	// Get the row pertaining to this UID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "DOC"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("invoiceTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed retrieving document with UID %s. Error %s", UID, err.Error())
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}

	docJSON := row.Columns[2].GetBytes()
	docPDF := row.Columns[3].GetBytes()
	currStatus := row.Columns[4].GetString_()

	//Start- Check that the currentStatus to newStatus transition is accurate

	stateTransitionAllowed := false

	//SUBMITTED_BY_EB -> ACCEPTED_BY_IB
	//SUBMITTED_BY_EB -> REJECTED_BY_IB

	if currStatus == "SUBMITTED_BY_EB" && newStatus == "ACCEPTED_BY_IB" {
		stateTransitionAllowed = true
	} else if currStatus == "SUBMITTED_BY_EB" && newStatus == "REJECTED_BY_IB" {
		stateTransitionAllowed = true
	}

	if stateTransitionAllowed == false {
		return nil, errors.New("This state transition is not allowed.")
	}

	//End- Check that the currentStatus to newStatus transition is accurate

	err = stub.DeleteRow(
		"invoiceTable",
		columns,
	)
	if err != nil {
		return nil, errors.New("Failed deleting row.")
	}

	_, err = stub.InsertRow(
		"invoiceTable",
		shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: "DOC"}},
				&shim.Column{Value: &shim.Column_String_{String_: UID}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: docJSON}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: docPDF}},
				&shim.Column{Value: &shim.Column_String_{String_: newStatus}}},
		})
	if err != nil {
		return nil, errors.New("Failed inserting row.")
	}

	return nil, nil

}

// GetJSON () – returns as JSON a single document w.r.t. the UID
func (t *Invoice) GetJSON(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	// Get the row pertaining to this UID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "DOC"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("invoiceTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed retrieving document with UID %s. Error %s", UID, err.Error())
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}

	return row.Columns[2].GetBytes(), nil

}

// GetPDF () – returns as JSON a single document w.r.t. the UID
func (t *Invoice) GetPDF(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	// Get the row pertaining to this UID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "DOC"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("invoiceTable", columns)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed retrieveing document with UID " + UID + ". Error " + err.Error() + ". \"}"
		return nil, errors.New(jsonResp)
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}

	return row.Columns[3].GetBytes(), nil
}

// GetStatus () – returns as JSON the Status w.r.t. the UID
func (t *Invoice) GetStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	// Get the row pertaining to this UID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "DOC"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("invoiceTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed retrieving document with UID %s. Error %s", UID, err.Error())
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}

	return []byte(row.Columns[4].GetString_()), nil
}
