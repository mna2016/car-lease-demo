package main

import (
	"errors"
	"fmt"
	"time"
	"bytes"
	"unicode/utf8"

//	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)

var logger = shim.NewLogger("CLDChaincode")

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const REGULATOR			=  "REG"
const   AF      		=  "AF "
const   DMA     		=  "DMA"
const   SUPPLIER      	=  "SUP"
const   TRANSPORTER     =  "TRP"

const   AUTHORITY      =  "regulator"
const   MANUFACTURER   =  "manufacturer"
const   PRIVATE_ENTITY =  "private"
const   LEASE_COMPANY  =  "lease_company"
const   SCRAP_MERCHANT =  "scrap_merchant"


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can
//					be done to the vehicle at points in it's lifecycle
//==============================================================================================================================
const   STATE_TEMPLATE  			=  0
const   STATE_MANUFACTURE  			=  1
const   STATE_PRIVATE_OWNERSHIP 	=  2
const   STATE_LEASED_OUT 			=  3
const   STATE_BEING_SCRAPPED  		=  4

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

/* Vehicle structure commented out 
//==============================================================================================================================
//	Vehicle - Defines the structure for a *** object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Vehicle struct {
	Make            string `json:"make"`
	Model           string `json:"model"`
	Reg             string `json:"reg"`
	VIN             int    `json:"VIN"`
	Owner           string `json:"owner"`
	Scrapped        bool   `json:"scrapped"`
	Status          int    `json:"status"`
	Colour          string `json:"colour"`
	V5cID           string `json:"v5cID"`
	LeaseContractID string `json:"leaseContractID"`
}
*/



//BEGIN new vehicle data structure
type Vehicle struct {

		TransactionType 	string	`json:"transactionType"`
		OwnerId				string	`json:"ownerId"`
		AssetId				string	`json:"assetId"`
		MatnrAf				string	`json:"matnrAf"`
		PoDma				string	`json:"poDma"`
		PoSupp				string	`json:"poSupp"`
		DmaDelDate			string	`json:"dmaDelDate"`
		AfDelDate			string	`json:"afDelDate"`
		TruckMod			string	`json:"truckMod"`
		TruckPDate			string	`json:"truckPdate"`
		TruckChnum			string	`json:"truckChnum"`
		TruckEnnum			string	`json:"truckEnnum"`
		SuppTest			string	`json:"suppTest"`
		GrDma				string	`json:"grDma"`
		GrAf				string	`json:"grAf"`
		DmaMasdat			string	`json:"dmaMasdat"`
		AfDmaTest			string	`json:"afDmaTest"`
		DmaDelCert			string	`json:"dmaDelCert"`
		AfDoc				string	`json:"afDoc"`
		Caller				string  `json:"caller"`
		V5cID           string `json:"v5cID"`

		}  

//END new vehicle Data Structure



//Animal struct for local storage of input arguments read from incoming transaction
type Animal struct {
		TransactionType 	string	`json:"transactionType"`
		OwnerId				string	`json:"ownerId"`
		AssetId				string	`json:"assetID"`		//BLOCKCHAIN KEY
															//NOTE: assetId changed to assetID based on request from UI developer
		MatnrAf				string	`json:"matnrAf"`
		PoDma				string	`json:"poDma"`
		PoSupp				string	`json:"poSupp"`
		DmaDelDate			string	`json:"dmaDelDate"`
		AfDelDate			string	`json:"afDelDate"`
		TruckMod			string	`json:"truckMod"`
		TruckPDate			string	`json:"truckPdate"`
		TruckChnum			string	`json:"truckChnum"`
		TruckEnnum			string	`json:"truckEnnum"`
		SuppTest			string	`json:"suppTest"`
		GrDma				string	`json:"grDma"`
		GrAf				string	`json:"grAf"`
		DmaMasdat			string	`json:"dmaMasdat"`
		AfDmaTest			string	`json:"afDmaTest"`
		DmaDelCert			string	`json:"dmaDelCert"`
		AfDoc				string	`json:"afDoc"`
		Caller				string  `json:"caller"`		//the UI/person who fired the transaction
		V5cid           string `json:"v5cID"`

		}


/* momentary structure to hol input json*/
type InRequest struct {
		Asset struct {	
		
		TransactionType 	string	`json:"transactionType"`
		OwnerId				string	`json:"ownerId"`
		AssetId				string	`json:"assetID"`		//BLOCKCHAIN KEY
															//NOTE: assetId changed to assetID based on request from UI developer
		MatnrAf				string	`json:"matnrAf"`
		PoDma				string	`json:"poDma"`
		PoSupp				string	`json:"poSupp"`
		DmaDelDate			string	`json:"dmaDelDate"`
		AfDelDate			string	`json:"afDelDate"`
		TruckMod			string	`json:"truckMod"`
		TruckPDate			string	`json:"truckPdate"`
		TruckChnum			string	`json:"truckChnum"`
		TruckEnnum			string	`json:"truckEnnum"`
		SuppTest			string	`json:"suppTest"`
		GrDma				string	`json:"grDma"`
		GrAf				string	`json:"grAf"`
		DmaMasdat			string	`json:"dmaMasdat"`
		AfDmaTest			string	`json:"afDmaTest"`
		DmaDelCert			string	`json:"dmaDelCert"`
		AfDoc				string	`json:"afDoc"`
		Caller				string  `json:"caller"`		//the UI/person who fired the transaction
		V5cid           string `json:"v5cID"`
		} 
}
		
//==============================================================================================================================
//	V5C Holder - Defines the structure that holds all the v5cIDs for vehicles that have been created.
//				Used as an index when querying all vehicles.
//==============================================================================================================================

type V5C_Holder struct {
	V5Cs 	[]string `json:"v5cs"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert string `json:"ecert"`
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	//Args
	//				0
	//			peer_address

	var v5cIDs V5C_Holder

	bytes, err := json.Marshal(v5cIDs)

    if err != nil { return nil, errors.New("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

	for i:=0; i < len(args); i=i+2 {
		t.add_ecert(stub, args[i], args[i+1])
	}

	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 get_ecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) get_ecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil { return nil, errors.New("Couldn't retrieve ecert for user " + name) }

	return ecert, nil
}

//==============================================================================================================================
//	 add_ecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) add_ecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {


	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
    affiliation, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	return string(affiliation), nil

}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

    // if err != nil { return "", "", err }

	// ecert, err := t.get_ecert(stub, user);

    // if err != nil { return "", "", err }

	affiliation, err := t.check_affiliation(stub);

    if err != nil { return "", "", err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrieve_v5c - Gets the state of the data at v5cID in the ledger then converts it from the stored
//					JSON into the Vehicle struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_v5c(stub shim.ChaincodeStubInterface, v5cID string) (Vehicle, error) {

	var v Vehicle

	bytes, err := stub.GetState(v5cID);

	if err != nil {	fmt.Printf("RETRIEVE_V5C: Failed to invoke vehicle_code: %s", err); return v, errors.New("RETRIEVE_V5C: Error retrieving vehicle with v5cID = " + v5cID) }

	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("RETRIEVE_V5C: Corrupt vehicle record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_V5C: Corrupt vehicle record"+string(bytes))	}

	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the Vehicle struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v Vehicle) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting vehicle record: %s", err); return false, errors.New("Error converting vehicle record") }

	err = stub.PutState(v.V5cID, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing vehicle record: %s", err); return false, errors.New("Error storing vehicle record") }

	return true, nil
}



//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub)

	//if err != nil { return nil, errors.New("Error retrieving caller information")}
    if err != nil { fmt.Printf("Error retrieving caller information")}

    
    // Get the args from the transaction proposal
    Args := stub.GetStringArgs()
    //byte1,err1:= stub.GetArgsSlice()
    
    //byte1:= stub.GetArgs()
    
    //if (err1 != nil) {fmt.Printf("No error")}
     // if len(args) != 2 {
     //   return nil("Incorrect arguments. Expecting a key and a value")
      //}
    
    var animals Animal
    var inreq InRequest
	//err2 := json.Unmarshal([]byte(Args[1]), &animals)
	err2 := json.Unmarshal([]byte(Args[1]), &inreq)
	if err2 != nil {
		fmt.Println("error:", err2)
	}
    
	//Now assign the real arguments to object animal 
	animals = inreq.Asset
	
	//return nil, errors.New("INPUT REQUEST:" + Args[1]+ "animals struct :"+ animals.AssetId + ":" + animals.Caller + ":" )
	// IMPORTANT: v5cid variable is used in most of the places in this contract
	// the frontend will pass the field assetID as the
	// copy assetID over to v5cid here
	animals.V5cid = animals.AssetId
	
    //fmt.Println("Input Arguments are: %v", animals)
    //return nil, errors.New("animals struct :"+ animals.V5cid + ":" + animals.Make + ":")

	if function == "create_vehicle" {
        return t.create_vehicle(stub, "DVLA", AUTHORITY, animals.V5cid)
	} else if function == "createAsset" {
		return t.createAsset(stub, "DVLA", AUTHORITY, animals.V5cid)
    } else if function == "ping" {
        return t.ping(stub)
    } else { 																				// If the function is not a create then there must be a *** so we need to retrieve the ***.
		//argPos := 1
		

		if function == "scrap_vehicle" {																// If its a scrap vehicle then only two arguments are passed (no update value) all others have three arguments and the v5cID is expected in the last argument
			//argPos = 0
			fmt.Printf("scrap vehicle")
		}

		//v, err := t.retrieve_v5c(stub, args[argPos])
		v, err := t.retrieve_v5c(stub, animals.V5cid)

        if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); return nil, errors.New("Error retrieving v5c") }


        if strings.Contains(function, "update") == true {


				/*if 		   function == "authority_to_manufacturer" { return t.authority_to_manufacturer(stub, v, caller, caller_affiliation, args[0], "manufacturer")
				} else if  function == "manufacturer_to_private"   { return t.manufacturer_to_private(stub, v, caller, caller_affiliation, args[0], "private")
				} else if  function == "private_to_private" 	   { return t.private_to_private(stub, v, caller, caller_affiliation, args[0], "private")
				} else if  function == "private_to_lease_company"  { return t.private_to_lease_company(stub, v, caller, caller_affiliation, args[0], "lease_company")
				} else if  function == "lease_company_to_private"  { return t.lease_company_to_private(stub, v, caller, caller_affiliation, args[0], "private")
				} else if  function == "private_to_scrap_merchant" { return t.private_to_scrap_merchant(stub, v, caller, caller_affiliation, args[0], "scrap_merchant")
				}

		} else if function == "update_make"  	    { return t.update_make(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_model"        { return t.update_model(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_reg" { return t.update_registration(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_vin" 			{ return t.update_vin(stub, v, caller, caller_affiliation, args[0])
        } else if function == "update_colour" 		{ return t.update_colour(stub, v, caller, caller_affiliation, args[0])
		} else if function == "scrap_vehicle" 		{ return t.scrap_vehicle(stub, v, caller, caller_affiliation) 
		} else*/ 
				if function == "updateAsset" 		{ return t.updateAsset(stub, v, caller, caller_affiliation,"dummy new value",animals) 
				} else if function == "updateDoc" 		{ return t.updateDoc(stub, v, caller, caller_affiliation,animals) }
		
	}
	}
	return nil, errors.New("Function of the name "+ function +" doesn't exist.")

}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.


//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub)
	//if err != nil { fmt.Printf("QUERY: Error retrieving caller details", err); return nil, errors.New("QUERY: Error retrieving caller details: "+err.Error()) }
    if err != nil { fmt.Printf("Error retrieving caller information")}
    logger.Debug("function: ", function)
    logger.Debug("caller: ", caller)
    logger.Debug("affiliation: ", caller_affiliation)
    // Get the args from the transaction proposal
    Args := stub.GetStringArgs()
    var animals Animal
    var inreq InRequest
	//err2 := json.Unmarshal([]byte(Args[1]), &animals)
	err2 := json.Unmarshal([]byte(Args[1]), &inreq)
	if err2 != nil {
		fmt.Println("error:", err2)
	}
	
	
	animals = inreq.Asset
    
	// IMPORTANT: v5cid variable is used in most of the places in this contract
	// the frontend will pass the field assetID as the
	// copy assetID over to v5cid here
	animals.V5cid = animals.AssetId
	
	caller = animals.Caller
    caller_affiliation = AUTHORITY
    
    

	if function == "get_vehicle_details" || function == "readAsset"  {
		if len(args) != 1 { fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed") }
		v, err := t.retrieve_v5c(stub, animals.V5cid)
		if err != nil { fmt.Printf("QUERY: Error retrieving v5c: %s", err); return nil, errors.New("QUERY: Error retrieving v5c "+err.Error())}
		return t.get_vehicle_details(stub, v, caller, caller_affiliation)
		 
	} else if function == "check_unique_v5c" {
		return t.check_unique_v5c(stub, animals.V5cid, caller, caller_affiliation)
	} else if function == "get_vehicles" || function == "readAllAssets" {
		return t.get_vehicles(stub, caller, caller_affiliation)
	} else if function == "get_ecert" {
		return t.get_ecert(stub,animals.V5cid)
	} else if function == "ping" {
		return t.ping(stub)
	} else if function == "readDoc" {
		 v, err := t.retrieve_v5c(stub, animals.V5cid)
		 _ = err
		 return t.readDoc(stub, v, caller, caller_affiliation) 
	}

	return nil, errors.New("Received unknown function invocation " + function)

}

//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Vehicle - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_vehicle(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, v5cID string) ([]byte, error) {
	var v Vehicle

	v5c_ID         := "\"v5cID\":\""+v5cID+"\", "							// Variables to define the JSON
	vin            := "\"VIN\":0, "
	make           := "\"Make\":\"UNDEFINED\", "
	model          := "\"Model\":\"UNDEFINED\", "
	reg            := "\"Reg\":\"UNDEFINED\", "
	owner          := "\"Owner\":\""+caller+"\", "
	colour         := "\"Colour\":\"UNDEFINED\", "
	leaseContract  := "\"LeaseContractID\":\"UNDEFINED\", "
	status         := "\"Status\":0, "
	scrapped       := "\"Scrapped\":false"

	vehicle_json := "{"+v5c_ID+vin+make+model+reg+owner+colour+leaseContract+status+scrapped+"}" 	// Concatenates the variables to create the total JSON object

	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(v5cID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits

												if err != nil { fmt.Printf("CREATE_VEHICLE: Invalid v5cID: %s", err); return nil, errors.New("Invalid v5cID") }

	if 				v5c_ID  == "" 	 ||
					matched == false    {
																		fmt.Printf("CREATE_VEHICLE: Invalid v5cID provided");
																		return nil, errors.New("Invalid v5cID provided=>"+v5cID+"<")
	}

	err = json.Unmarshal([]byte(vehicle_json), &v)							// Convert the JSON defined above into a vehicle object for go

																		if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(v.V5cID) 								// If not an error then a record exists so cant create a new *** with this V5cID as it must be unique

																		if record != nil { return nil, errors.New("Vehicle already exists") }

	if 	caller_affiliation != AUTHORITY {							// Only the regulator can create a new v5c

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_vehicle. %v === %v", caller_affiliation, AUTHORITY))

	}

	_, err  = t.save_changes(stub, v)

																		if err != nil { fmt.Printf("CREATE_VEHICLE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("v5cIDs")

																		if err != nil { return nil, errors.New("Unable to get v5cIDs") }

	var v5cIDs V5C_Holder

	err = json.Unmarshal(bytes, &v5cIDs)

																		if err != nil {	return nil, errors.New("Corrupt V5C_Holder record") }

	v5cIDs.V5Cs = append(v5cIDs.V5Cs, v5cID)


	bytes, err = json.Marshal(v5cIDs)

															if err != nil { fmt.Print("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

															if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

//=================================================================================================================================
//	 Create Asset - Creates the initial JSON for the asset and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) createAsset(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, v5cID string) ([]byte, error) {
	var v Vehicle

	//Initialize variables which will make up the JSON structure that will be written to world state
		TransactionType 	:= "\"transactionType\":\"UNDEFINED\", "
		OwnerId				:= "\"ownerId\":\""+REGULATOR+"\", "			//owner at the time of creation is always regulator
		AssetId				:= "\"assetID\":\""+v5cID+"\", "				//NOTE:assetId changed to assetID based on UI developer request
		MatnrAf				:= "\"matnrAf\":\"UNDEFINED\", "
		PoDma				:= "\"poDma\":\"UNDEFINED\", "
		PoSupp				:= "\"poSupp\":\"UNDEFINED\", "
		DmaDelDate			:= "\"dmaDelDate\":\"UNDEFINED\", "
		AfDelDate			:= "\"afDelDate\":\"UNDEFINED\", "
		TruckMod			:= "\"truckMod\":\"UNDEFINED\", "
		TruckPDate			:= "\"truckPdate\":\"UNDEFINED\", "
		TruckChnum			:= "\"truckChnum\":\"UNDEFINED\", "
		TruckEnnum			:= "\"truckEnnum\":\"UNDEFINED\", "
		SuppTest			:= "\"suppTest\":\"UNDEFINED\", "
		GrDma				:= "\"grDma\":\"UNDEFINED\", "
		GrAf				:= "\"grAf\":\"UNDEFINED\", "
		DmaMasdat			:= "\"dmaMasdat\":\"UNDEFINED\", "
		AfDmaTest			:= "\"afDmaTest\":\"UNDEFINED\", "
		DmaDelCert			:= "\"dmaDelCert\":\"UNDEFINED\", "
		AfDoc				:= "\"afDoc\":\"UNDEFINED\", "
		Caller				:= "\"caller\":\"UNDEFINED\" "
		v5c_ID         		:= "\"v5cID\":\""+v5cID+"\", "							// Variables to define the JSON



	vehicle_json := "{"+v5c_ID+TransactionType+OwnerId+AssetId+MatnrAf+PoDma+PoSupp+DmaDelDate+AfDelDate+TruckMod+TruckPDate+TruckChnum+TruckEnnum+SuppTest+GrDma+GrAf+DmaMasdat+AfDmaTest+DmaDelCert+AfDoc+Caller+"}" 	// Concatenates the variables to create the total JSON object

	//matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(v5cID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits
	
	//NOTE: format check changed as per request from SAP and UI developers
	// NOW matched = true if the v5cID passed fits format: 10 numeric digits
	matched, err := regexp.Match("^[0-9]{10}", []byte(v5cID))  				
		
												if err != nil { fmt.Printf("CREATE_VEHICLE: Invalid v5cID: %s", err); return nil, errors.New("Invalid v5cID") }

	if 				v5c_ID  == "" 	 ||
					matched == false    {
																		fmt.Printf("CREATE_VEHICLE: Invalid v5cID provided");
																		return nil, errors.New("Invalid v5cID provided=>"+v5cID+"<")
	}

	err = json.Unmarshal([]byte(vehicle_json), &v)							// Convert the JSON defined above into a vehicle object for go

																		if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(v.V5cID) 								// If not an error then a record exists so cant create a new *** with this V5cID as it must be unique

																		if record != nil { return nil, errors.New("Vehicle already exists") }

	if 	caller_affiliation != AUTHORITY {							// Only the regulator can create a new v5c

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_vehicle. %v === %v", caller_affiliation, AUTHORITY))

	}

	_, err  = t.save_changes(stub, v)

																		if err != nil { fmt.Printf("CREATE_VEHICLE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("v5cIDs")

																		if err != nil { return nil, errors.New("Unable to get v5cIDs") }

	var v5cIDs V5C_Holder

	err = json.Unmarshal(bytes, &v5cIDs)

																		if err != nil {	return nil, errors.New("Corrupt V5C_Holder record") }

	v5cIDs.V5Cs = append(v5cIDs.V5Cs, v5cID)


	bytes, err = json.Marshal(v5cIDs)

															if err != nil { fmt.Print("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

															if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

//=================================================================================================================================
//	 Update Asset - Updates the initial JSON for the asset and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) updateAsset(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string, animals Animal) ([]byte, error) {

//if the transaction is fired by a person who owns this asset then he has the right to update
	if 	v.OwnerId == animals.Caller		{
				

					if 	animals.TransactionType			!= ""	{ v.TransactionType = animals.TransactionType 	} 	
					if	animals.OwnerId 				!= ""	{ v.OwnerId = animals.OwnerId					} 
					
					if	animals.MatnrAf					!= "" 	{ v.MatnrAf = animals.MatnrAf					}
					if	animals.PoDma					!= ""	{ v.PoDma = animals.PoDma						}
					if	animals.PoSupp					!= ""	{ v.PoSupp = animals.PoSupp						}
					if	animals.DmaDelDate				!= ""	{ v.DmaDelDate = animals.DmaDelDate				}	
					if	animals.AfDelDate				!= "" 	{ v.AfDelDate = animals.AfDelDate				}
					if	animals.TruckMod				!= ""	{ v.TruckMod = animals.TruckMod					}
					if	animals.TruckPDate				!= ""	{ v.TruckPDate = animals.TruckPDate				}
					if	animals.TruckChnum				!= ""	{ v.TruckChnum = animals.TruckChnum				}	
					if	animals.TruckEnnum				!= ""	{ v.TruckEnnum = animals.TruckEnnum				}	
					if	animals.SuppTest				!= "" 	{ v.SuppTest = animals.SuppTest					}
					if	animals.GrDma					!= ""	{ v.GrDma = animals.GrDma						}
					if	animals.GrAf					!= "" 	{ v.GrAf = animals.GrAf							}
					if	animals.DmaMasdat				!= ""	{ v.DmaMasdat = animals.DmaMasdat				}	
					if	animals.AfDmaTest				!= ""	{ v.AfDmaTest = animals.AfDmaTest				}
					if 	animals.DmaDelCert				!= "" 	{ v.DmaDelCert = animals.DmaDelCert				}	
					if	animals.AfDoc					!= "" 	{ v.AfDoc = animals.AfDoc						}
					//if	animals.Make					!= "" 	{ v.Make = animals.Make							}            	
					if 	animals.Caller					!= ""	{ v.Caller = animals.Caller						}


			} else {

		return nil, errors.New(fmt.Sprint("Permission denied. updateAsset %t %t" + v.OwnerId == caller, caller_affiliation == MANUFACTURER))
	}
			v.AssetId = v.V5cID	//assetId and v5cid are the same thing. 					

	_, err := t.save_changes(stub, v)

		if err != nil { fmt.Printf("updateAsset: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}


/* COMMENTED the vehicle "update" functions as of 7/26/2017 6:00PM

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 authority_to_manufacturer
//=================================================================================================================================
func (t *SimpleChaincode) authority_to_manufacturer(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if     	v.Status				== STATE_TEMPLATE	&&
			v.Owner					== caller			&&
			caller_affiliation		== AUTHORITY		&&
			recipient_affiliation	== MANUFACTURER		&&
			v.Scrapped				== false			{		// If the roles and users are ok

					v.Owner  = recipient_name		// then make the owner the new owner
					v.Status = STATE_MANUFACTURE			// and mark it in the state of manufacture

	} else {									// Otherwise if there is an error
															fmt.Printf("AUTHORITY_TO_MANUFACTURER: Permission Denied");
                                                            return nil, errors.New(fmt.Sprintf("Permission Denied. authority_to_manufacturer. %v %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v, v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))


	}

	_, err := t.save_changes(stub, v)						// Write new state

															if err != nil {	fmt.Printf("AUTHORITY_TO_MANUFACTURER: Error saving changes: %s", err); return nil, errors.New("Error saving changes")	}

	return nil, nil									// We are Done

}

//=================================================================================================================================
//	 manufacturer_to_private
//=================================================================================================================================
func (t *SimpleChaincode) manufacturer_to_private(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if 		v.Make 	 == "UNDEFINED" ||
			v.Model  == "UNDEFINED" ||
			v.Reg 	 == "UNDEFINED" ||
			v.Colour == "UNDEFINED" ||
			v.VIN == 0				{					//If any part of the *** is undefined it has not bene fully manufacturered so cannot be sent
															fmt.Printf("MANUFACTURER_TO_PRIVATE: Car not fully defined")
															return nil, errors.New(fmt.Sprintf("Car not fully defined. %v", v))
	}

	if 		v.Status				== STATE_MANUFACTURE	&&
			v.Owner					== caller				&&
			caller_affiliation		== MANUFACTURER			&&
			recipient_affiliation	== PRIVATE_ENTITY		&&
			v.Scrapped     == false							{

					v.Owner = recipient_name
					v.Status = STATE_PRIVATE_OWNERSHIP

	} else {
        return nil, errors.New(fmt.Sprintf("Permission Denied. manufacturer_to_private. %v %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v, v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))
    }

	_, err := t.save_changes(stub, v)

	if err != nil { fmt.Printf("MANUFACTURER_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 private_to_private
//=================================================================================================================================
func (t *SimpleChaincode) private_to_private(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if 		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner					== caller					&&
			caller_affiliation		== PRIVATE_ENTITY			&&
			recipient_affiliation	== PRIVATE_ENTITY			&&
			v.Scrapped				== false					{

					v.Owner = recipient_name

	} else {
        return nil, errors.New(fmt.Sprintf("Permission Denied. private_to_private. %v %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v, v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))
	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("PRIVATE_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 private_to_lease_company
//=================================================================================================================================
func (t *SimpleChaincode) private_to_lease_company(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if 		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner					== caller					&&
			caller_affiliation		== PRIVATE_ENTITY			&&
			recipient_affiliation	== LEASE_COMPANY			&&
            v.Scrapped     			== false					{

					v.Owner = recipient_name

	} else {
        return nil, errors.New(fmt.Sprintf("Permission denied. private_to_lease_company. %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))

	}

	_, err := t.save_changes(stub, v)
															if err != nil { fmt.Printf("PRIVATE_TO_LEASE_COMPANY: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 lease_company_to_private
//=================================================================================================================================
func (t *SimpleChaincode) lease_company_to_private(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner  				== caller					&&
			caller_affiliation		== LEASE_COMPANY			&&
			recipient_affiliation	== PRIVATE_ENTITY			&&
			v.Scrapped				== false					{

				v.Owner = recipient_name

	} else {
		return nil, errors.New(fmt.Sprintf("Permission Denied. lease_company_to_private. %v %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v, v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))
	}

	_, err := t.save_changes(stub, v)
															if err != nil { fmt.Printf("LEASE_COMPANY_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 private_to_scrap_merchant
//=================================================================================================================================
func (t *SimpleChaincode) private_to_scrap_merchant(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, recipient_name string, recipient_affiliation string) ([]byte, error) {

	if		v.Status				== STATE_PRIVATE_OWNERSHIP	&&
			v.Owner					== caller					&&
			caller_affiliation		== PRIVATE_ENTITY			&&
			recipient_affiliation	== SCRAP_MERCHANT			&&
			v.Scrapped				== false					{

					v.Owner = recipient_name
					v.Status = STATE_BEING_SCRAPPED

	} else {
        return nil, errors.New(fmt.Sprintf("Permission Denied. private_to_scrap_merchant. %v %v === %v, %v === %v, %v === %v, %v === %v, %v === %v", v, v.Status, STATE_PRIVATE_OWNERSHIP, v.Owner, caller, caller_affiliation, PRIVATE_ENTITY, recipient_affiliation, SCRAP_MERCHANT, v.Scrapped, false))
	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("PRIVATE_TO_SCRAP_MERCHANT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 Update Functions
//=================================================================================================================================
//	 update_vin
//=================================================================================================================================
func (t *SimpleChaincode) update_vin(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	new_vin, err := strconv.Atoi(string(new_value)) 		                // will return an error if the new vin contains non numerical chars

															if err != nil || len(string(new_value)) != 15 { return nil, errors.New("Invalid value passed for new VIN") }

	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&&
			caller_affiliation	== MANUFACTURER			&&
			v.VIN				== 0					&&			// Can't change the VIN after its initial assignment
			v.Scrapped			== false				{

					v.VIN = new_vin					// Update to the new value
	} else {

        return nil, errors.New(fmt.Sprintf("Permission denied. update_vin %v %v %v %v %v", v.Status, STATE_MANUFACTURE, v.Owner, caller, v.VIN, v.Scrapped))

	}

	_, err  = t.save_changes(stub, v)						// Save the changes in the blockchain

															if err != nil { fmt.Printf("UPDATE_VIN: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}


//=================================================================================================================================
//	 update_registration
//=================================================================================================================================
func (t *SimpleChaincode) update_registration(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string) ([]byte, error) {


	if		v.Owner				== caller			&&
			caller_affiliation	!= SCRAP_MERCHANT	&&
			v.Scrapped			== false			{

					v.Reg = new_value

	} else {
        return nil, errors.New(fmt.Sprint("Permission denied. update_registration"))
	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("UPDATE_REGISTRATION: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_colour
//=================================================================================================================================
func (t *SimpleChaincode) update_colour(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 		v.Owner				== caller				&&
			caller_affiliation	== MANUFACTURER			&&/*((v.Owner				== caller			&&
			caller_affiliation	== MANUFACTURER)		||
			caller_affiliation	== AUTHORITY)			&&*/
/*			v.Scrapped			== false				{

					v.Colour = new_value
	} else {

		return nil, errors.New(fmt.Sprint("Permission denied. update_colour %t %t %t" + v.Owner == caller, caller_affiliation == MANUFACTURER, v.Scrapped))
	}

	_, err := t.save_changes(stub, v)

		if err != nil { fmt.Printf("UPDATE_COLOUR: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_make
//=================================================================================================================================
func (t *SimpleChaincode) update_make(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&&
			caller_affiliation	== MANUFACTURER			&&
			v.Scrapped			== false				{

					v.Make = new_value
	} else {

        return nil, errors.New(fmt.Sprint("Permission denied. update_make %t %t %t" + v.Owner == caller, caller_affiliation == MANUFACTURER, v.Scrapped))


	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("UPDATE_MAKE: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_model
//=================================================================================================================================
func (t *SimpleChaincode) update_model(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, new_value string) ([]byte, error) {

	if 		v.Status			== STATE_MANUFACTURE	&&
			v.Owner				== caller				&&
			caller_affiliation	== MANUFACTURER			&&
			v.Scrapped			== false				{

					v.Model = new_value

	} else {
        return nil, errors.New(fmt.Sprint("Permission denied. update_model %t %t %t" + v.Owner == caller, caller_affiliation == MANUFACTURER, v.Scrapped))

	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("UPDATE_MODEL: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 scrap_vehicle
//=================================================================================================================================
func (t *SimpleChaincode) scrap_vehicle(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	if		v.Status			== STATE_BEING_SCRAPPED	&&
			v.Owner				== caller				&&
			caller_affiliation	== SCRAP_MERCHANT		&&
			v.Scrapped			== false				{

					v.Scrapped = true

	} else {
		return nil, errors.New("Permission denied. scrap_vehicle")
	}

	_, err := t.save_changes(stub, v)

															if err != nil { fmt.Printf("SCRAP_VEHICLE: Error saving changes: %s", err); return nil, errors.New("SCRAP_VEHICLError saving changes") }

	return nil, nil

}
*/
//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_vehicle_details
//=================================================================================================================================
func (t *SimpleChaincode) get_vehicle_details(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	bytes1, err := json.Marshal(v)
	
	
	//make up the response in such a manner to have the response sandwiched
	//between mspart1 and msgpart2. this is done to reformat the response 
	//which the UI expects. The data stored on the blockchain will have no change 
	//in structure. only the vehicle struct will be stored on blockchain. 
	//This reformat is just an adjustment.

	var str bytes.Buffer
	
	str1 := string(bytes1)
	msgpart1  := "{\"assetstate\":{\"asset\":"
	

	txnID := stub.GetTxID()
	txntmsp,errN := stub.GetTxTimestamp()
	
	out1 := "2017-07-13T14:23:45.262161407Z"
	timetemp1 := time.Unix(txntmsp.Seconds, int64(txntmsp.Nanos))
	time1 := timetemp1.Format(out1)	
	//time1 := timetemp.String()
	_ = errN
	
	
    
	msgpart2  := "},\"txnid\":\""
	msgpart3  := "\",\"txnts\":\""
	msgpart4  := "\"}" //txnid and txnts to be populated
	
	str.WriteString(msgpart1)
	str.WriteString(str1)
	str.WriteString(msgpart2)
	str.WriteString(txnID)
	str.WriteString(msgpart3)
	str.WriteString(time1)
	str.WriteString(msgpart4)
	
	
	bytes3 := []byte(str.String())

	
	
	if err != nil { return nil, errors.New("READASSET: Invalid vehicle object") }

	if 		v.OwnerId	== caller		||
				caller  == REGULATOR	{

					//return bytes, nil
					return bytes3, nil
	} else {
				return nil, errors.New("Permission Denied. readAsset. The caller should be owner or Regulator.")
	}

}


func (t *SimpleChaincode) get_vehicle_details2(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {

	bytes1, err := json.Marshal(v)
	
	//make up the response in such a manner to have the response sandwiched
	//between mspart1 and msgpart2. this is done to reformat the response 
	//which the UI expects. The data stored on the blockchain will have no change 
	//in structure. only the vehicle struct will be stored on blockchain. 
	//This reformat is just an adjustment.

	var str bytes.Buffer
	
	str1 := string(bytes1)
	/*
	msgpart1  := "{\"assetstate\":{{\"asset\":"
	msgpart2  := "}},\"txnid\":\"\",\"txnts\":\"\"}" //txnid and txnts to be populated

	str.WriteString(msgpart1)
	str.WriteString(str1)
	str.WriteString(msgpart2)
	*/
	msgpart1  := "{\"assetstate\":{\"asset\":"
	

	txnID := stub.GetTxID()
	txntmsp,errN := stub.GetTxTimestamp()
	time1 := time.Unix(txntmsp.Seconds, int64(txntmsp.Nanos)).String()
	_ = errN
	
	msgpart2  := "},\"txnid\":\""
	msgpart3  := "\",\"txnts\":\""
	msgpart4  := "\"}" //txnid and txnts to be populated
	
	str.WriteString(msgpart1)
	str.WriteString(str1)
	str.WriteString(msgpart2)
	str.WriteString(txnID)
	str.WriteString(msgpart3)
	str.WriteString(time1)
	str.WriteString(msgpart4)
	
	
	bytes3 := []byte(str.String())
	
	
	if err != nil { return nil, errors.New("READASSET: Invalid vehicle object") }

	if 		v.OwnerId	== caller		||
				caller  == REGULATOR	{

					//return bytes, nil
					return bytes3, nil
	} else {
				return nil, errors.New("Permission Denied. readAsset. The caller should be owner or Regulator.")
	}

}

//=================================================================================================================================
//	 get_vehicles
//=================================================================================================================================

func (t *SimpleChaincode) get_vehicles(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string) ([]byte, error) {

	if 	caller  != REGULATOR	{

 			return nil, errors.New("Permission Denied! Only a REGULATOR may read list of all assets")
	}


	bytes, err := stub.GetState("v5cIDs")

	if err != nil { return nil, errors.New("Unable to get v5cIDs or assetIds") }

	var v5cIDs V5C_Holder

	err = json.Unmarshal(bytes, &v5cIDs)

	if err != nil {	return nil, errors.New("Corrupt V5C_Holder or AssetID Holder") }

	result := "["

	var temp []byte
	var v Vehicle

	for _, v5c := range v5cIDs.V5Cs {

		v, err = t.retrieve_v5c(stub, v5c)

		if err != nil {return nil, errors.New("Failed to retrieve V5C or AssetId")}

		temp, err = t.get_vehicle_details2(stub, v, caller, caller_affiliation)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 check_unique_v5c
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_v5c(stub shim.ChaincodeStubInterface, v5c string, caller string, caller_affiliation string) ([]byte, error) {
	_, err := t.retrieve_v5c(stub, v5c)
	if err == nil {
		return []byte("false"), errors.New("V5C or AssetId is not unique")
	} else {
		return []byte("true"), nil
	}
}

//=================================================================================================================================
//	 Update Doc - Attaches document to a blockchain block
//=================================================================================================================================
func (t *SimpleChaincode) updateDoc(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string, animals Animal) ([]byte, error) {

//if the transaction is fired by a person who owns this asset then he has the right to update
	if 	v.OwnerId == animals.Caller		{
					
					if	animals.AfDoc					== "" 	{ return nil, errors.New("AfDoc cannot be empty when updateDoc is called!")}

					if  utf8.RuneCountInString(animals.AfDoc) > 250000 { return nil, errors.New("AfDoc cannot be larger than 250KB!")}

					} else {

		return nil, errors.New(fmt.Sprint("Permission denied. updateAsset %t %t" + v.OwnerId == caller, caller_affiliation == MANUFACTURER))
	}
	
	v.AssetId = v.V5cID	//assetId and v5cid are the same thing. 					
	v.AfDoc = animals.AfDoc //move input to vehicle structure

	//Now post the document to blockchain
	_, err := t.save_changes(stub, v)

		if err != nil { fmt.Printf("updateAsset: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 Update Doc - Attaches document to a blockchain block
//=================================================================================================================================
func (t *SimpleChaincode) readDoc(stub shim.ChaincodeStubInterface, v Vehicle, caller string, caller_affiliation string) ([]byte, error) {


	//if the transaction is fired by a person who owns this asset then he has the right to update
	//if 	v.OwnerId == animals.Caller		{

	bytes1, err := json.Marshal(v.AfDoc)
	
	var str bytes.Buffer
	
	str1 := string(bytes1)
		msgpart1  := "{\"assetstate\":{\"asset\":"
	

	txnID := stub.GetTxID()
	txntmsp,errN := stub.GetTxTimestamp()
	time1 := time.Unix(txntmsp.Seconds, int64(txntmsp.Nanos)).String()
	_ = errN
	
	msgpart2  := "},\"txnid\":\""
	msgpart3  := "\",\"txnts\":\""
	msgpart4  := "\"}" //txnid and txnts to be populated
	
	str.WriteString(msgpart1)
	str.WriteString(str1)
	str.WriteString(msgpart2)
	str.WriteString(txnID)
	str.WriteString(msgpart3)
	str.WriteString(time1)
	str.WriteString(msgpart4)
	

	bytes3 := []byte(str.String())

	
	
	if err != nil { return nil, errors.New("READASSET: Invalid vehicle object") }

	if 		v.OwnerId	== caller		||
				caller  == REGULATOR	{

					//return bytes, nil
					return bytes3, nil
	} else {
				return nil, errors.New("Permission Denied. readAsset. The caller should be owner or Regulator.")
	}



}



//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))

	if err != nil { fmt.Printf("Error starting Chaincode: %s", err) }
}
