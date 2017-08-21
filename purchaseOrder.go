package main

import (
	"encoding/json"
	"bytes"
	"errors"
	"fmt"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)


//ALL_PO key to refer the purchaseOrder master data
const ALL_PO = "ALL_PO"
var logger = shim.NewLogger("PurchaseOrder")

type PurchaseOrder struct {
	RefNo string
	Importer string
	Exporter	string
	Commodity	string
	Aircompressor	string
	Currency	string
	UnitPrice	string
	Amount	string
	Quantity	string
	Weight string
	TermsofPayment string
	TermsofTrade string
	TermsofInsurance string
	PackingMethod string
	WayofTransportation string
	TimeofShipment string
	PortofShipment string
	PortofDischarge string
	PaymentDate string
	PORejectReason string
}

//Init initializes the document smart contract
func (t *PurchaseOrder) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
		//Place an empty arry
	stub.PutState(ALL_PO, []byte("[]"))
	stub.PutState("id", []byte("1"))
	return nil, nil
}

// Creating a new Purchase Order
func(t *PurchaseOrder) createPO(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	payload := args[0]
	who := args[1]
	fmt.Println("new Payload is " + payload)
	//validate new po
	valMsg := t.validatePO(who, payload)
	// for getting uniqueId, this'll give new id per second
	 poNo:= "1"//time.Now().Local().Format("20060102150405")
	//If there is no error messages then create the UFA	
	if valMsg == "" {
		stub.PutState(poNo, []byte(payload))
		fmt.Println("new poNo is " + poNo)
		//t.updateMasterRecords(stub, poNo)
			logger.Info("Created the PO after successful validation : " + payload)
	} else {
		return nil, errors.New("Validation failure: " + valMsg)
	}
	return nil, nil
}

//Validate a PO
func (t *PurchaseOrder) validatePO(who string, payload string) string {

	//As of now I am checking if who is of proper role
	var validationMessage bytes.Buffer
	var ufaDetails map[string]string

	logger.Info("validateNewPO")
	
	if who == "IMPORTER BANK" {
		json.Unmarshal([]byte(payload), &ufaDetails)
		if ufaDetails["Currency"] != "Rs"{
			logger.Info(ufaDetails["Currency"])
			validationMessage.WriteString("\naIncorrect PurchaseOrder")
		}
		//Now check individual fields
		
	} else {
		validationMessage.WriteString("\naAccess Denied to create a PO")
	}
	logger.Info("Validation messagge " + validationMessage.String())
	logger.Info(ufaDetails["Currency"])
	return validationMessage.String()
}

//Append a newPO number to the master list
func (t *PurchaseOrder) updateMasterRecords(stub shim.ChaincodeStubInterface, poNo string) error {
	var recordList []string
	recBytes, _ := stub.GetState(ALL_PO)

	err := json.Unmarshal(recBytes, &recordList)
	if err != nil {
		return errors.New("Failed to unmarshal updateMasterReords ")
	}
	recordList = append(recordList, poNo)
	bytesToStore, _ := json.Marshal(recordList)
	logger.Info("After addition" + string(bytesToStore))
	stub.PutState(ALL_PO, bytesToStore)
	return nil
}
//get all the newPo
func (t *PurchaseOrder) getAllPo(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("getAllPo called")
	var allApp []map[string]string
	recordsList, err := getAllRecordsList(stub)
	if err != nil {
		return nil, errors.New("Unable to get all the records ")
	}
	for _, value := range recordsList {
		recBytes, _ := t.getPoDetails(stub, value)

		var record map[string]string
		json.Unmarshal(recBytes, &record)
		allApp = append(allApp, record)
	}
	outputBytes, _ := json.Marshal(allApp)
	logger.Info("Returning records from getAllPo " + string(outputBytes))
	return outputBytes, nil
}
//Returns all the Po Numbers stored
func getAllRecordsList(stub shim.ChaincodeStubInterface) ([]string, error) {
	var recordList []string
	recBytes, _ := stub.GetState(ALL_PO)

	err := json.Unmarshal(recBytes, &recordList)
	if err != nil {
		return nil, errors.New("Failed to unmarshal getAllRecordsList ")
	}

	return recordList, nil
}
//Get a single PO
func (t *PurchaseOrder) getPoDetails(stub shim.ChaincodeStubInterface, args string) ([]byte, error) {
	logger.Info("getPoDetails called with PO number: " + args)
	var jsonResp string
	poNumber := args //PO num
	//who :=args[1] //Role
	recBytes, err := stub.GetState(poNumber)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + poNumber + "\"}"
		return nil, errors.New(jsonResp)
	}
	if recBytes == nil{
		jsonResp = "{\"Message\":\"No record exists for " + poNumber + "\"}"
		return []byte(jsonResp),nil
	
	}
	logger.Info("Returning records from getUFADetails " + string(recBytes))
	return recBytes, nil
}


