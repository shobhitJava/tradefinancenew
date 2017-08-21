package main

// Rahul Hundet 26-07-2017
// - validation for number of argumants 10 in submitLC as client code could not pass cert arguments, so save certs as blank
// - Removed logging related stuff as the package could not be found on bluemix service
// - Hardcoded Certs to blank in SubmitLC

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	//logging "github.com/op/go-logging"
)

// Access control flag - perform access control if flag is true
//change to false to test
const accessControlFlag bool = false

//var myLogger = logging.MustGetLogger("access_control_helper")

// Contract struct
type Contract struct {
	ContractID     string `json:"contractID"`
	ContractStatus string `json:"contractStatus"`
	Comment        string `json:"comment`
}

type POJSON struct {
	UID              string `json:"UID"`
	Status           string `json:"Status"`
	ImporterName     string `json:"ImporterName"`
	ExporterName     string `json:"ExporterName"`
	ImporterBankName string `json:"ImporterBankName"`
	ExporterBankName string `json:"ExporterBankName"`
	ImporterCert     []byte `json:"ImporterCert"`
	ExporterCert     []byte `json:"ExporterCert"`
	ImporterBankCert []byte `json:"ImporterBankCert"`
	ExporterBankCert []byte `json:"ExporterBankCert"`
	ShippingCompany  string `json:"ShippingCompany"`
	InsuranceCompany string `json:"InsuranceCompany"`
}

// ContractsList struct
type ContractsList struct {
	Contracts []Contract `json:"contracts"`
}

// Participants struct
//type Participants struct {
//	ImporterBankName string `json:"importerBankName"`
//	ExporterBankName string `json:"exporterBankName"`
//}

// Participant struct
type Participant struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// ParticipantList struct
type ParticipantList struct {
	Participants []Participant `json:"participants"`
}

//ResultJSON ...
type ResultJSON struct {
	ContractID       string
	ImporterCert     []byte
	ImporterBankCert []byte
	ExporterBankCert []byte
	ExporterCert     []byte
}

// TF is a high level smart contract that TFs together business artifact based smart contracts
type TF struct {
	lc      LC
	bl      BL
	invoice Invoice
	pl      PL
	po      PurchaseOrder
}

// Init initializes the smart contracts
//Fabric version migration to 0.6
//func (t *TF) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *TF) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	// Check if table already exists
	_, err := stub.GetTable("BPTable")
	if err == nil {
		// Table already exists; do not recreate
		return nil, nil
	}

	// Create Business Process Table
	err = stub.CreateTable("BPTable", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Type", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "UID", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Status", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "ImporterName", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "ExporterName", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "ImporterBankName", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "ExporterBankName", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "ImporterCert", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "ExporterCert", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "ImporterBankCert", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "ExporterBankCert", Type: shim.ColumnDefinition_BYTES, Key: false},
		&shim.ColumnDefinition{Name: "ShippingCompany", Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: "InsuranceCompany", Type: shim.ColumnDefinition_STRING, Key: false},
	})
	if err != nil {
		return nil, errors.New("Failed creating BPTable.")
	}

	t.lc.Init(stub, function, args)
	t.bl.Init(stub, function, args)
	t.invoice.Init(stub, function, args)
	t.pl.Init(stub, function, args)
	t.po.Init(stub, function, args)

	return nil, nil
}

// isCaller is a helper function that verifies the signature of the caller given the certificate to match with

//Fabric version migration to 0.6
//func (t *TF) isCaller(stub *shim.ChaincodeStub, certificate []byte) (bool, error) {
func (t *TF) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	//	myLogger.Debugf("Check caller...")
	fmt.Printf("PDD-DBG: Check caller...")

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed getting metadata")
	}
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed getting payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed getting binding")
	}

	////	myLogger.Debugf("passed certificate [% x]", certificate)
	//	myLogger.Debugf("passed sigma [% x]", sigma)
	//	myLogger.Debugf("passed payload [% x]", payload)
	//	myLogger.Debugf("passed binding [% x]", binding)

	fmt.Printf("PDD-DBG: passed certificate [% x]", certificate)
	fmt.Printf("PDD-DBG: passed sigma [% x]", sigma)
	fmt.Printf("PDD-DBG: passed payload [% x]", payload)
	fmt.Printf("PDD-DBG: passed binding [% x]", binding)

	ok, err := stub.VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		//		myLogger.Error("Failed checking signature ", err.Error())
		fmt.Printf("PDD-DBG: Failed checking signature %s", err.Error())
		return ok, err
	}
	if !ok {
		//		myLogger.Error("Invalid signature")
		fmt.Printf("PDD-DBG: Invalid signature")
	}

	//myLogger.Debug("Check caller...Verified!")
	//fmt.Printf("PDD-DBG: Check caller...Verified!")

	return ok, err
}

// isCallerImporter accepts UID as input and checks if the caller is importer Bank
//Fabric version migration to 0.6
//func (t *TF) isCallerImporter(stub *shim.ChaincodeStub, args []string) (bool, error) {
func (t *TF) isCallerImporter(stub shim.ChaincodeStubInterface, args []string) (bool, error) {
	if len(args) != 1 {
		return false, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return false, errors.New("Failed retrieving row with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return false, errors.New("Failed retrieving row with contract ID " + UID)
	}

	// Get the importer bank's certificate for this contract - 5th column in the table
	certificate := row.Columns[7].GetBytes()

	ok, err := t.isCaller(stub, certificate)
	if err != nil {
		return false, errors.New("Failed checking importer bank's identity")
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

// ExporterBank accepts UID as input and checks if the caller is Exporter Bank
//Fabric version migration to 0.6
//func (t *TF) isCallerExporter(stub *shim.ChaincodeStub, args []string) (bool, error) {
func (t *TF) isCallerExporter(stub shim.ChaincodeStubInterface, args []string) (bool, error) {
	if len(args) != 1 {
		return false, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return false, errors.New("Failed retrieving row with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return false, errors.New("Failed retrieving row with contract ID " + UID)
	}

	// Get the exporter bank's certificate for this contract - 6th column in the table
	certificate := row.Columns[8].GetBytes()

	ok, err := t.isCaller(stub, certificate)
	if err != nil {
		return false, errors.New("Failed checking exporter bank's identity " + err.Error())
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

// isCallerImporterBank accepts UID as input and checks if the caller is importer Bank
//Fabric version migration to 0.6
//func (t *TF) isCallerImporterBank(stub *shim.ChaincodeStub, args []string) (bool, error) {
func (t *TF) isCallerImporterBank(stub shim.ChaincodeStubInterface, args []string) (bool, error) {
	if len(args) != 1 {
		return false, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return false, errors.New("Failed retrieving row with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return false, errors.New("Failed retrieving row with contract ID " + UID)
	}

	// Get the importer bank's certificate for this contract - 5th column in the table
	certificate := row.Columns[9].GetBytes()

	ok, err := t.isCaller(stub, certificate)
	if err != nil {
		return false, errors.New("Failed checking importer bank's identity")
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

// ExporterBank accepts UID as input and checks if the caller is Exporter Bank
//Fabric version migration to 0.6
//func (t *TF) isCallerExporterBank(stub *shim.ChaincodeStub, args []string) (bool, error) {
func (t *TF) isCallerExporterBank(stub shim.ChaincodeStubInterface, args []string) (bool, error) {

	if len(args) != 1 {
		return false, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return false, errors.New("Failed retrieving row with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return false, errors.New("Failed retrieving row with contract ID " + UID)
	}

	// Get the exporter bank's certificate for this contract - 6th column in the table
	certificate := row.Columns[10].GetBytes()

	ok, err := t.isCaller(stub, certificate)
	if err != nil {
		return false, errors.New("Failed checking exporter bank's identity " + err.Error())
	}
	if !ok {
		return false, nil
	}

	return true, nil
}

// isCallerParticipant accepts UID as input and checks if the caller is Exporter Bank
//Fabric version migration to 0.6
//func (t *TF) isCallerParticipant(stub *shim.ChaincodeStub, args []string) (bool, error) {
func (t *TF) isCallerParticipant(stub shim.ChaincodeStubInterface, args []string) (bool, error) {
	if len(args) != 1 {
		return false, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return false, errors.New("Failed retrieving row with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return false, errors.New("Failed retrieving row with contract ID " + UID)
	}

	// Get certificates
	certificate1 := row.Columns[7].GetBytes()
	certificate2 := row.Columns[8].GetBytes()
	certificate3 := row.Columns[9].GetBytes()
	certificate4 := row.Columns[10].GetBytes()

	ok1, err1 := t.isCaller(stub, certificate1)
	ok2, err2 := t.isCaller(stub, certificate2)
	ok3, err3 := t.isCaller(stub, certificate3)
	ok4, err4 := t.isCaller(stub, certificate4)

	if err1 != nil && err2 != nil && err3 != nil && err4 != nil {
		return false, errors.New(err1.Error() + " " + err2.Error() + " " + err3.Error() + " " + err4.Error())
	}

	if !ok1 && !ok2 && !ok3 && !ok4 {
		return false, nil
	}

	return true, nil
}

func (t *TF) GetBPJSON(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed retrieving document with ContractNo %s. Error %s", UID, err.Error())
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}
	var poJSON POJSON

	poJSON.UID = row.Columns[1].GetString_()
	poJSON.Status = row.Columns[2].GetString_()
	poJSON.ImporterName = row.Columns[3].GetString_()
	poJSON.ExporterName = row.Columns[4].GetString_()
	poJSON.ImporterBankName = row.Columns[5].GetString_()
	poJSON.ExporterBankName = row.Columns[6].GetString_()
	poJSON.ImporterCert = row.Columns[7].GetBytes()
	poJSON.ExporterCert = row.Columns[8].GetBytes()
	poJSON.ImporterBankCert = row.Columns[9].GetBytes()
	poJSON.ExporterBankCert = row.Columns[10].GetBytes()
	poJSON.ShippingCompany = row.Columns[11].GetString_()
	poJSON.InsuranceCompany = row.Columns[12].GetString_()

	jsonPO, err := json.Marshal(poJSON)

	if err != nil {

		return nil, err
	}

	fmt.Println(jsonPO)

	return jsonPO, nil

}

// getContractCerts is a function built for testing to retrieve the certificates stored for a contract
func (t *TF) getContractCerts(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	UID := args[0]

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return nil, errors.New("Failed retrieving row in BPTable with contract ID " + UID + ". Error " + err.Error())
	}
	if len(row.Columns) == 0 {
		return nil, errors.New("Failed retrieving row in BPTable with contract ID " + UID)
	}

	var res ResultJSON
	res.ContractID = row.Columns[1].GetString_()
	res.ImporterCert = row.Columns[7].GetBytes()
	res.ImporterBankCert = row.Columns[9].GetBytes()
	res.ExporterBankCert = row.Columns[10].GetBytes()
	res.ExporterCert = row.Columns[8].GetBytes()

	resjson, err := json.Marshal(res)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal json result")
	}

	return []byte(resjson), nil

}

// getNumContracts get total number of LC applications. Helper function to generate next contract ID.
//Fabric version migration to 0.6
//func (t *TF) getNumContracts(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) getNumContracts(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0.")
	}

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)

	contractCounter := 0

	rows, err := stub.GetRows("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve row")
	}

	for row := range rows {
		if len(row.Columns) != 0 {
			contractCounter++
		}
	}

	type count struct {
		NumContracts int
	}

	var c count
	c.NumContracts = contractCounter

	return json.Marshal(c)
}

// listContracts  lists all the contracts
//Fabric version migration to 0.6
//func (t *TF) listContracts(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) listContracts(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0.")
	}

	var allContractsList ContractsList

	// Get the row pertaining to this contractID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)

	rows, err := stub.GetRows("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve row")
	}

	allContractsList.Contracts = make([]Contract, 0)

	for row := range rows {
		if len(row.Columns) == 0 {
			res := make(map[string]string, 0)
			resjson, err := json.Marshal(res)
			return resjson, err
		}

		var nextContract Contract
		nextContract.ContractID = row.Columns[1].GetString_()

		b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
		if err != nil {
			return nil, err
		}

		if string(b) == "ACCEPTED_BY_EB" {

			b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
			if string(b1) == "" {
				nextContract.ContractStatus = string(b)
			} else {

				nextContract.ContractStatus = string(b1)
			}

		} else {

			nextContract.ContractStatus = string(b)
			nextContract.Comment = string(c)
		}

		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
			if err != nil {
				return nil, err
			}
			if res == true {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}
		} else {
			allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
		}

	}

	return json.Marshal(allContractsList)
}

func (t *TF) listContractsByRoleName(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2.")
	}

	var allContractsList ContractsList

	companyID := args[0]
	roleID := args[1]

	if roleID == "4" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		for row := range rows {
			if len(row.Columns) == 0 {
				break
			}

			var nextContract Contract

			if row.Columns[3].GetString_() == companyID {
				nextContract.ContractID = row.Columns[1].GetString_()
				if nextContract.ContractID != "" {
					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}
				}
			}
		}
	}

	if roleID == "1" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		//var contractIDOfUser ContractsList1

		for row := range rows {

			//contractIDOfUser.ContractNo = ""

			if len(row.Columns) == 0 {

				break

			}

			var nextContract Contract

			if row.Columns[4].GetString_() == companyID {

				nextContract.ContractID = row.Columns[1].GetString_()

				if nextContract.ContractID != "" {

					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}

				}

			}
		}

	}

	if roleID == "5" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		for row := range rows {

			if len(row.Columns) == 0 {

				break

			}

			var nextContract Contract

			if row.Columns[5].GetString_() == companyID {

				nextContract.ContractID = row.Columns[1].GetString_()

				if nextContract.ContractID != "" {

					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}

				}

			}

		}

	}

	if roleID == "2" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		for row := range rows {

			if len(row.Columns) == 0 {

				break

			}

			var nextContract Contract

			if row.Columns[6].GetString_() == companyID {

				nextContract.ContractID = row.Columns[1].GetString_()

				if nextContract.ContractID != "" {

					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}

				}

			}

		}

	}

	if roleID == "3" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		for row := range rows {

			if len(row.Columns) == 0 {

				break

			}

			var nextContract Contract

			if row.Columns[11].GetString_() == companyID {

				nextContract.ContractID = row.Columns[1].GetString_()

				if nextContract.ContractID != "" {

					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}

				}

			}

		}

	}

	if roleID == "6" {

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)

		rows, err := stub.GetRows("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve row")
		}

		allContractsList.Contracts = make([]Contract, 0)

		//var contractIDOfUser ContractsList1

		for row := range rows {

			//contractIDOfUser.ContractNo = ""

			if len(row.Columns) == 0 {

				break

			}

			var nextContract Contract

			if row.Columns[12].GetString_() == companyID {

				nextContract.ContractID = row.Columns[1].GetString_()

				if nextContract.ContractID != "" {

					b, c, err := t.lc.GetStatus(stub, []string{nextContract.ContractID})
					if err != nil {
						return nil, err
					}

					if string(b) == "ACCEPTED_BY_EB" {

						b1, _ := t.bl.GetStatus(stub, []string{nextContract.ContractID})
						if string(b1) == "" {
							nextContract.ContractStatus = string(b)
						} else {

							nextContract.ContractStatus = string(b1)
						}

					} else {

						nextContract.ContractStatus = string(b)
						nextContract.Comment = string(c)
					}

					if accessControlFlag == true {
						res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
						if err != nil {
							return nil, err
						}
						if res == true {
							allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
						}
					} else {
						allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
					}

				}

			}

		}

	}

	return json.Marshal(allContractsList)

}

// listContractsByRole  lists all the contracts where the user belongs to the provided role.
//Fabric version migration to 0.6
//func (t *TF) listContractsByRole(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) listContractsByRole(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	var allContractsList ContractsList

	role := args[0]

	if role != "Importer" && role != "Exporter" && role != "ImporterBank" && role != "ExporterBank" {
		return nil, errors.New("Role should be Importer, Exporter, ImporterBank or ExporterBank.")
	}

	// Get the row pertaining to this contractID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)

	rows, err := stub.GetRows("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve row")
	}

	allContractsList.Contracts = make([]Contract, 0)

	for row := range rows {
		if len(row.Columns) == 0 {
			res := make(map[string]string, 0)
			resjson, err := json.Marshal(res)
			return resjson, err
		}

		var nextContract Contract
		nextContract.ContractID = row.Columns[1].GetString_()

		if role == "Importer" && accessControlFlag == true {
			res, err := t.isCallerImporter(stub, []string{nextContract.ContractID})
			if err != nil {
				return nil, err
			}
			if res == true {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		} else if role == "Exporter" && accessControlFlag == true {
			res, err := t.isCallerExporter(stub, []string{nextContract.ContractID})
			if err != nil {
				return nil, err
			}
			if res == true {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		} else if role == "ImporterBank" && accessControlFlag == true {
			res, err := t.isCallerImporterBank(stub, []string{nextContract.ContractID})
			if err != nil {
				return nil, err
			}
			if res == true {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		} else if role == "ExporterBank" && accessControlFlag == true {
			res, err := t.isCallerExporterBank(stub, []string{nextContract.ContractID})
			if err != nil {
				return nil, err
			}
			if res == true {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		} else {
			allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
		}

	}

	return json.Marshal(allContractsList)
}

//listLCsByStatus  lists all the contracts
//Fabric version migration to 0.6
//func (t *TF) listLCsByStatus(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) listLCsByStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	status := args[0]
	var allContractsList ContractsList

	// Get the row pertaining to this contractID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)

	rows, err := stub.GetRows("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve row")
	}

	allContractsList.Contracts = make([]Contract, 0)

	for row := range rows {
		// GetRows returns empty message if key does not exist
		if len(row.Columns) == 0 {
			res := make(map[string]string, 0)
			resjson, err := json.Marshal(res)
			return resjson, err
		}

		var nextContract Contract

		b, _, err := t.lc.GetStatus(stub, []string{row.Columns[1].GetString_()})
		if err != nil {
			return nil, err
		}

		if status == string(b) {
			nextContract.ContractID = row.Columns[1].GetString_()
			//allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			if accessControlFlag == true {
				res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
				if err != nil {
					return nil, err
				}
				if res == true {
					allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
				}
			} else {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		}

	}

	return json.Marshal(allContractsList)
}

//listEDsByStatus  lists all the contracts

//Fabric version migration to 0.6
//func (t *TF) listEDsByStatus(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) listEDsByStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	status := args[0]
	//docType := args[1]

	//if docType != "BL" && docType != "INVOICE" && docType != "PACKINGLIST" {
	//	return nil, errors.New("Document type should be BL or INVOICE or PACKINGLIST.")
	//}

	var allContractsList ContractsList

	// Get the row pertaining to this contractID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)

	rows, err := stub.GetRows("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve row")
	}

	allContractsList.Contracts = make([]Contract, 0)

	for row := range rows {
		// GetRows returns empty message if key does not exist
		if len(row.Columns) == 0 {
			res := make(map[string]string, 0)
			resjson, err := json.Marshal(res)
			return resjson, err
		}

		var nextContract Contract

		//since all export documents are always kept in the same state, it is enough to check against one.
		b, err := t.bl.GetStatus(stub, []string{row.Columns[1].GetString_()})
		if err != nil {
			return nil, err
		}
		if status == string(b) {
			nextContract.ContractID = row.Columns[1].GetString_()
			//allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			if accessControlFlag == true {
				res, err := t.isCallerParticipant(stub, []string{nextContract.ContractID})
				if err != nil {
					return nil, err
				}
				if res == true {
					allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
				}
			} else {
				allContractsList.Contracts = append(allContractsList.Contracts, nextContract)
			}

		}

	}

	return json.Marshal(allContractsList)
}

// getContractParticipants () – returns as JSON the Status w.r.t. the UID
//Fabric version migration to 0.6
//func (t *TF) getContractParticipants(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
func (t *TF) getContractParticipants(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1.")
	}

	var participantList ParticipantList
	participantList.Participants = make([]Participant, 0)

	UID := args[0]

	// Get the row pertaining to this UID
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: UID}}
	columns = append(columns, col2)

	row, err := stub.GetRow("BPTable", columns)
	if err != nil {
		return nil, fmt.Errorf("Error: Failed retrieving document with UID %s. Error %s", UID, err.Error())
	}

	// GetRows returns empty message if key does not exist
	if len(row.Columns) == 0 {
		return nil, nil
	}

	var participant Participant
	participant.ID = row.Columns[3].GetString_()
	participant.Role = "Importer"
	participantList.Participants = append(participantList.Participants, participant)

	participant.ID = row.Columns[4].GetString_()
	participant.Role = "Exporter"
	participantList.Participants = append(participantList.Participants, participant)

	participant.ID = row.Columns[5].GetString_()
	participant.Role = "ImporterBank"
	participantList.Participants = append(participantList.Participants, participant)

	participant.ID = row.Columns[6].GetString_()
	participant.Role = "ExporterBank"
	participantList.Participants = append(participantList.Participants, participant)

	return json.Marshal(participantList.Participants)
}

// crossCheckDocs() is a helper function that checks if the submitted documents are consistent with each other
func (t *TF) crossCheckDocs(args []string) (bool, error) {

	if len(args) != 4 {
		return false, errors.New("Incorrect number of arguments. Expecting 4.")
	}

	lcJSON := []byte(args[0])
	blJSON := []byte(args[1])
	invoiceJSON := []byte(args[2])
	packingListJSON := []byte(args[3])

	var lc LC
	var bl BL
	var invoice Invoice
	var pl PL

	err := json.Unmarshal(lcJSON, &lc)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(blJSON, &bl)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(invoiceJSON, &invoice)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(packingListJSON, &pl)
	if err != nil {
		return false, err
	}

	if lc.Tag20 != bl.LC_NUMBER || lc.Tag20 != invoice.LC_NUMBER || lc.Tag20 != pl.DOCUMENTARY_CREDIT_NUMBER {
		return false, errors.New("LC numbers on all documents do not match each other")
	}

	return true, nil
}

// Invoke invokes the chaincode
//Fabric version migration to 0.6
//func (t *TF) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *TF) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "submitLC" {
		/*
			if len(args) != 10 {
				return nil, fmt.Errorf("Incorrect number of arguments. Expecting 10. Got: %d.", len(args))
			}
		*/

		UID := args[0]
		fmt.Println(UID)
		lcJSON := args[1]
		fmt.Println(lcJSON)
		importerName := args[2]
		exporterName := args[3]
		importerBankName := args[4]
		exporterBankName := args[5]
		/*
			importerCert := []byte(args[6])
			exporterCert := []byte(args[7])
			importerBankCert := []byte(args[8])
			exporterBankCert := []byte(args[9])
		*/
		// Hardcoded certs to blank
		importerCert := []byte("")
		exporterCert := []byte("")
		importerBankCert := []byte("")
		exporterBankCert := []byte("")

		shippingCompany := ""
		insuranceCompany := ""

		// Insert a row
		ok, err := stub.InsertRow("BPTable", shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: "BP"}},
				&shim.Column{Value: &shim.Column_String_{String_: UID}},
				&shim.Column{Value: &shim.Column_String_{String_: "STARTED"}},
				&shim.Column{Value: &shim.Column_String_{String_: importerName}},
				&shim.Column{Value: &shim.Column_String_{String_: exporterName}},
				&shim.Column{Value: &shim.Column_String_{String_: importerBankName}},
				&shim.Column{Value: &shim.Column_String_{String_: exporterBankName}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: importerCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: exporterCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: importerBankCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: exporterBankCert}},
				&shim.Column{Value: &shim.Column_String_{String_: shippingCompany}},
				&shim.Column{Value: &shim.Column_String_{String_: insuranceCompany}}},
		})

		if err != nil {
			return nil, err
		}
		if !ok && err == nil {
			return nil, errors.New("Row already exists.")
		}

		return t.lc.SubmitDoc(stub, []string{UID, lcJSON, ""})
	} else if function == "acceptLC" {
		if accessControlFlag == true {
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}
		args = append(args, "ACCEPTED_BY_EB")
		return t.lc.UpdateStatus(stub, args)
	} else if function == "paymentReceived" {

		if accessControlFlag == true {
			//res, err := t.isCallerExporterBank(stub, []string{args[0], string(sigma), string(payload), string(binding)})
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		lcStatus, _, err := t.lc.GetStatus(stub, []string{args[0]})
		if err != nil {
			return nil, err
		}

		if string(lcStatus) == "PAYMENT_DUE_FROM_IB_TO_EB" {
			args = append(args, "Payment")
			args = append(args, "PAYMENT_RECEIVED")
			return t.lc.UpdateStatus(stub, args)
		}
		return nil, errors.New("Payment is not yet due.")

	} else if function == "defaultedOnPayment" {
		if accessControlFlag == true {
			//res, err := t.isCallerExporterBank(stub, []string{args[0], string(sigma), string(payload), string(binding)})
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		args = append(args, "Payment_defaulted")
		args = append(args, "PAYMENT_DEFAULTED")

		return t.lc.UpdateStatus(stub, args)
	} else if function == "rejectLC" {
		if accessControlFlag == true {
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}
		args = append(args, "REJECTED_BY_EB")
		return t.lc.UpdateStatus(stub, args)
	} else if function == "reSubmitLC" {
		if len(args) != 11 {
			return nil, fmt.Errorf("Incorrect number of arguments. Expecting 11. Got: %d.", len(args))
		}

		UID := args[0]
		lcJSON := args[1]
		comment := args[10]

		return t.lc.ReSubmitDoc(stub, []string{UID, lcJSON, "", comment})

	} else if function == "submitED" {
		/*if accessControlFlag == true {
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}*/

		if len(args) != 9 {
			return nil, fmt.Errorf("Incorrect number of arguments. Expecting 9. Got: %d.", len(args))
		}

		contractID := args[0]
		BLPDF := args[1]
		invoicePDF := args[2]
		packingListPDF := args[3]
		BLJSON := args[4]
		invoiceJSON := args[5]
		packingListJSON := args[6]
		shippingCompanyname := args[7]
		insuranceCompanyname := args[8]

		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: "BP"}}
		columns = append(columns, col1)
		col2 := shim.Column{Value: &shim.Column_String_{String_: contractID}}
		columns = append(columns, col2)

		row, err := stub.GetRow("BPTable", columns)
		if err != nil {
			return nil, fmt.Errorf("Error: Failed retrieving document with ContractNo %s. Error %s", contractID, err.Error())
		}

		// GetRows returns empty message if key does not exist
		if len(row.Columns) == 0 {
			return nil, nil
		}

		importerName := row.Columns[3].GetString_()
		exporterName := row.Columns[4].GetString_()
		importerBankName := row.Columns[5].GetString_()
		exporterBankName := row.Columns[6].GetString_()
		importerCert := row.Columns[7].GetBytes()
		exporterCert := row.Columns[8].GetBytes()
		importerBankCert := row.Columns[9].GetBytes()
		exporterBankCert := row.Columns[10].GetBytes()

		/*err = stub.DeleteRow(
			"BPTable",
			columns,
		)
		if err != nil {
			return nil, errors.New("Failed deleting row.")
		}
		*/

		ok, err := stub.ReplaceRow("BPTable", shim.Row{
			Columns: []*shim.Column{
				&shim.Column{Value: &shim.Column_String_{String_: "BP"}},
				&shim.Column{Value: &shim.Column_String_{String_: contractID}},
				&shim.Column{Value: &shim.Column_String_{String_: "STARTED"}},
				&shim.Column{Value: &shim.Column_String_{String_: importerName}},
				&shim.Column{Value: &shim.Column_String_{String_: exporterName}},
				&shim.Column{Value: &shim.Column_String_{String_: importerBankName}},
				&shim.Column{Value: &shim.Column_String_{String_: exporterBankName}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: importerCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: exporterCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: importerBankCert}},
				&shim.Column{Value: &shim.Column_Bytes{Bytes: exporterBankCert}},
				&shim.Column{Value: &shim.Column_String_{String_: shippingCompanyname}},
				&shim.Column{Value: &shim.Column_String_{String_: insuranceCompanyname}},
			},
		})

		if !ok && err == nil {

			return nil, errors.New("Document unable to Update.")
		}

		//Get the corresponding LC
		lcJSON, err := t.lc.GetJSON(stub, []string{contractID})
		if err != nil {
			return nil, err
		}

		//Validate that the BL is correct
		if BLJSON != string([]byte(`{}`)) {
			_, err = t.bl.ValidateDoc(stub, []string{BLJSON, string(lcJSON)})
			if err != nil {
				return nil, err
			}
		}

		//Validate that the invoice is correct
		if invoiceJSON != string([]byte(`{}`)) {
			_, err = t.invoice.ValidateDoc(stub, []string{invoiceJSON, string(lcJSON)})
			if err != nil {
				return nil, err
			}
		}

		//Validate that the packing list is correct
		if packingListJSON != string([]byte(`{}`)) {
			_, err = t.pl.ValidateDoc(stub, []string{packingListJSON, string(lcJSON)})
			if err != nil {
				return nil, err
			}
		}

		if BLJSON != string([]byte(`{}`)) && invoiceJSON != string([]byte(`{}`)) && packingListJSON != string([]byte(`{}`)) {
			res, err := t.crossCheckDocs([]string{string(lcJSON), string(BLJSON), string(invoiceJSON), string(packingListJSON)})
			if err != nil {
				return nil, err
			}

			if res == false {
				return nil, errors.New("Documents are not consistent with each other")
			}
		}

		//Submit the validated BL to the ledger
		if BLJSON != "" || BLPDF != "" {
			_, err = t.bl.SubmitDoc(stub, []string{contractID, BLJSON, BLPDF})
			if err != nil {
				return nil, err
			}
		}

		//Submit the validated invoice to the ledger
		if invoiceJSON != "" || invoicePDF != "" {
			_, err = t.invoice.SubmitDoc(stub, []string{contractID, invoiceJSON, invoicePDF})
			if err != nil {
				return nil, err
			}
		}

		//Submit the validated packing list to the ledger
		if packingListJSON != "" || packingListPDF != "" {
			_, err = t.pl.SubmitDoc(stub, []string{contractID, packingListJSON, packingListPDF})
			if err != nil {
				return nil, err
			}
		}

		//If pay on sight is true in letter of credit, do state transition LC:ACCEPTED -> PAYMENT_RECEIVED
		//var lc LC
		//err = json.Unmarshal(lcJSON, &lc)
		//if err != nil {
		//	return nil, err
		//}

		//if lc.Tag42C == "Sight" {
		//	return t.lc.UpdateStatus(stub, []string{contractID, "PAYMENT_RECEIVED"})
		//}
		/*
				args = append(args, "Payment_due")
			args = append(args, "PAYMENT_DUE_FROM_IB_TO_EB")
		*/
		_, err = t.lc.UpdateStatus(stub, args)

		return nil, nil
	} else if function == "acceptED" {

		if accessControlFlag == true {
			res, err := t.isCallerImporterBank(stub, []string{args[0]})

			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		//Get the corresponding LC
		lcJSON, err := t.lc.GetJSON(stub, []string{args[0]})
		if err != nil {
			return nil, err
		}

		var lc LC
		err = json.Unmarshal(lcJSON, &lc)
		if err != nil {
			return nil, err
		}

		/*if lc.Tag42C == "Sight" {

			t.lc.UpdateStatus(stub, []string{args[0],"Payment_Due", "PAYMENT_DUE_FROM_IB_TO_EB"})
		}*/

		args = append(args, "ACCEPTED_BY_IB")

		_, err = t.bl.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}
		_, err = t.invoice.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}
		_, err = t.pl.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}

		return nil, nil
	} else if function == "rejectED" {

		if accessControlFlag == true {

			res, err := t.isCallerImporterBank(stub, []string{args[0]})

			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		args = append(args, "REJECTED_BY_IB")

		_, err := t.bl.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}
		_, err = t.invoice.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}
		_, err = t.pl.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}

		return nil, nil
	} else if function == "acceptToPay" {

		if accessControlFlag == true {
			//res, err := t.isCallerImporterBank(stub, []string{args[0], string(sigma), string(payload), string(binding)})
			res, err := t.isCallerImporterBank(stub, []string{args[0]})

			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		args = append(args, "Payment_due")
		args = append(args, "PAYMENT_DUE_FROM_IB_TO_EB")

		_, err := t.lc.UpdateStatus(stub, args)
		if err != nil {
			return nil, err
		}

		return nil, nil
	} else if function == "createPO" {

		return t.po.createPO(stub, args)
	}

	return nil, errors.New("Invalid invoke function name.")
}

// Query callback representing the query of a chaincode
//Fabric version migration to 0.6
//func (t *TF) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
func (t *TF) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	type Status struct {
		Status string
	}
	status := Status{}

	type Result struct {
		Result string `json:"result"`
	}
	result := Result{}

	/*
		sigma, err := stub.GetCallerMetadata()
		if err != nil {
			return nil, errors.New("Failed getting metadata")
		}
		payload, err := stub.GetPayload()
		if err != nil {
			return nil, errors.New("Failed getting payload")
		}
		binding, err := stub.GetBinding()
		if err != nil {
			return nil, errors.New("Failed getting binding")
		}
	*/

	if function == "getLC" {

		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		return t.lc.GetJSON(stub, args)
	} else if function == "getBP" {

		return t.GetBPJSON(stub, args)

	} else if function == "getContractCerts" {

		return t.getContractCerts(stub, args)

	} else if function == "getLCStatus" {

		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		b, _, err := t.lc.GetStatus(stub, args)
		if err != nil {
			return nil, err
		}
		status.Status = string(b)

		return json.Marshal(status)
	} else if function == "validateLC" {

		b, err := t.lc.ValidateDoc(stub, args)
		if err != nil {
			return nil, err
		}
		result.Result = string(b)
		return json.Marshal(result)
	} else if function == "validateED" {
		if accessControlFlag == true {
			res, err := t.isCallerExporterBank(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments. Expecting 3.")
		}

		contractID := args[0]
		docType := args[1]
		docJSON := args[2]

		lcJSON, err := t.lc.GetJSON(stub, []string{contractID})
		if err != nil {
			return nil, err
		}

		if docType == "BL" {
			return t.bl.ValidateDoc(stub, []string{docJSON, string(lcJSON)})
		} else if docType == "INVOICE" {
			return t.invoice.ValidateDoc(stub, []string{docJSON, string(lcJSON)})
		} else if docType == "PACKINGLIST" {
			return t.pl.ValidateDoc(stub, []string{docJSON, string(lcJSON)})
		}

		return nil, nil
	} else if function == "getED" {
		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		if len(args) != 3 {
			return nil, errors.New("Incorrect number of arguments. Expecting 3.")
		}

		contractID := args[0]
		docType := args[1]
		docFormat := args[2]

		if docType != "BL" && docType != "INVOICE" && docType != "PACKINGLIST" {
			return nil, errors.New("Document type should be BL or INVOICE or PACKINGLIST")
		}

		if docFormat != "JSON" && docFormat != "PDF" {
			return nil, errors.New("Document format should be JSON or PDF")
		}

		if docFormat == "JSON" {
			if docType == "BL" {
				return t.bl.GetJSON(stub, []string{contractID})
			} else if docType == "INVOICE" {
				return t.invoice.GetJSON(stub, []string{contractID})
			} else if docType == "PACKINGLIST" {
				return t.pl.GetJSON(stub, []string{contractID})
			}

		} else if docFormat == "PDF" {
			if docType == "BL" {
				return t.bl.GetPDF(stub, []string{contractID})
			} else if docType == "INVOICE" {
				return t.invoice.GetPDF(stub, []string{contractID})
			} else if docType == "PACKINGLIST" {
				return t.pl.GetPDF(stub, []string{contractID})
			}

		}

		return nil, nil
	} else if function == "getEDStatus" {
		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		b, err := t.bl.GetStatus(stub, args)
		if err != nil {
			return nil, err
		}
		status.Status = string(b)
		return json.Marshal(status)
	} else if function == "getNumContracts" {

		return t.getNumContracts(stub, args)
	} else if function == "listContracts" {

		return t.listContracts(stub, args)
	} else if function == "listContractsByRole" {

		return t.listContractsByRole(stub, args)
	} else if function == "listContractsByRoleName" {

		return t.listContractsByRoleName(stub, args)
	} else if function == "listLCsByStatus" {

		return t.listLCsByStatus(stub, args)
	} else if function == "listEDsByStatus" {

		return t.listEDsByStatus(stub, args)
	} else if function == "getContractParticipants" {
		if accessControlFlag == true {
			res, err := t.isCallerParticipant(stub, []string{args[0]})
			if err != nil {
				return nil, err
			}
			if res == false {
				return nil, errors.New("Access denied.")
			}
		}

		return t.getContractParticipants(stub, args)
	} else if function == "isCallerExporterBank" {

		res, err := t.isCallerExporterBank(stub, args)

		if err != nil {
			return nil, err
		}
		if res == false {
			return nil, errors.New("Caller is not ExporterBank.")
		}

		if res == true {

			return []byte("true"), nil
		}

	} else if function == "getPoDetails" {
		return t.po.getPoDetails(stub, args[0])
	} else if function == "getAllPo" {
		return t.po.getAllPo(stub, args)
	}

	return nil, errors.New("Invalid query function name.")
}

func main() {
	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(TF))
	if err != nil {
		fmt.Printf("Error starting TF: %s", err)
	}
}
