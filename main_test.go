// Copyright 2019 Idan Dekel. All rights reserved.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

var (
	cert1        = []byte(`{"id":"1","title":"first cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This is the first certificate","transfer":{"to":"","status":""}}`)
	cert1Updated = []byte(`{"id":"1","title":"Updated cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This is the updated first certificate","transfer":{"to":"","status":""}}`)
	cert2        = []byte(`{"id":"2","title":"second cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This is the second certificate","transfer":{"to":"","status":""}}`)
	cert3        = []byte(`{"id":"3","title":"second cert","createdAt":"29 MAR 2019","ownerId":"11","year":2019,"note":"This is the third certificate","transfer":{"to":"","status":""}}`)
)

// IsEqualJSON performs a deep comparison on two JSONs, and returns an error if not equal
func IsEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	err := json.Unmarshal([]byte(s1), &o1)

	if err != nil {
		return false, err
	}

	err = json.Unmarshal([]byte(s2), &o2)

	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(o1, o2), nil
}

//executeRequest executes the right method, according to the path string
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/certificates/{id}", createCert).Methods("POST")
	router.HandleFunc("/certificates/{id}", updateCert).Methods("PUT")
	router.HandleFunc("/certificates/{id}", deleteCert).Methods("DELETE")

	router.HandleFunc("/users/{id}/certificates", listCerts).Methods("GET")

	router.HandleFunc("/certificates/{id}/transfers", createTransfer).Methods("POST")
	router.HandleFunc("/certificates/{id}/transfers", acceptTransfer).Methods("PUT")

	router.ServeHTTP(recorder, req)

	return recorder
}

// checkResponseCode verifies that the expected responce code has been received
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// TestCreateCertInvalidUser tries to create a certificate for a non-existing user. Verifies that it receives an error JSON
func TestCreateCertInvalidUser(t *testing.T) {

	cert := []byte(`{"id":"1","title":"Invalid User cert","createdAt":"29 MAR 2019","ownerId":"100","year":2019,"note":"This is a certificate created for an invalid user","transfer":{"to":"","status":""}}`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/1", bytes.NewBuffer(cert))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "User ID 100 is invalid. Cannot create certificate.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestCreate1stCert creates a certificate and checks the returned JSON to verify that it's been added to the certificates map
func TestCreate1stCert(t *testing.T) {

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/1", bytes.NewBuffer(cert1))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := `{"1":` + string(cert1) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("\nError code: %d\n", err)
	}
}

//TestUpdateCert updates the existing certificate, and verifies that the update has been saved to the certificates map
func TestUpdateCert(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/1", bytes.NewBuffer(cert1Updated))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := `{"1":` + string(cert1Updated) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("\nError code: %d\n", err)
	}
}

//TestUpdateCertInvalidID requests an update of a certificate with a non-existing ID. It then verifies that the update request has failed
func TestUpdateCertInvalidID(t *testing.T) {
	updatedCertInvalidID := []byte(`{"id":"11","title":"Updated cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This is the updated first certificate","transfer":{"to":"","status":""}}`)

	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/11", bytes.NewBuffer(updatedCertInvalidID))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Certificate ID 11 doesn't exist. Cannot update certificate.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestUpdateCertInvalidUserID requests an update of certificate 1 with a non-existing user ID. It then verifies that the update request has failed
func TestUpdateCertInvalidUserID(t *testing.T) {
	updatedCertInvalidUserID := []byte(`{"id":"1","title":"Updated cert","createdAt":"29 MAR 2019","ownerId":"100","year":2019,"note":"This is the updated first certificate","transfer":{"to":"","status":""}}`)

	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/1", bytes.NewBuffer(updatedCertInvalidUserID))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "User ID 100 is invalid. Cannot update certificate.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestCreate2ndCert is called after TestCreateCert. It creates a second certificate and verifies that it's been added to the certificates map
func TestCreate2ndCert(t *testing.T) {

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/2", bytes.NewBuffer(cert2))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := `{"1":` + string(cert1Updated) + `,"2":` + string(cert2) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("%v", err)
	}
}

//TestCreateCertWithExistingID creates a certificate with an ID that's already been used. Verifies that the certificate hasn't been added to the map
func TestCreateCertWithExistingID(t *testing.T) {

	cert := []byte(`{"id":"1","title":"Existing ID cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This certificate reuses an existing ID","transfer":{"to":"","status":""}}`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/1", bytes.NewBuffer(cert))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Certificate ID 1 already exists. Cannot create certificate.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestDeleteCertInvalidID tries to delete a certificate with a non-existing ID. It then verifies that the delete request has failed
func TestDeleteCertInvalidID(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8080/certificates/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Certificate ID 11 doesn't exist. Cannot delete certificate.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestDelete2ndCert send a delete request for the second certificate, and then verifies that it's been deleted from the certificates map
func TestDelete2ndCert(t *testing.T) {

	req, _ := http.NewRequest("DELETE", "http://localhost:8080/certificates/2", bytes.NewBuffer(cert2))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := `{"1":` + string(cert1Updated) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("\nError code: %d\n", err)
	}
}

//TestListCertsInvalidUser requests a list of certificates for an invalid user and verifies that it receives an error message
func TestListCertsInvalidUser(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/users/100/certificates", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
	expected := "User ID 100 is invalid. Cannot list certificates.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestListCertsEmptyList requests a list of certificates for a user with no certificates and verifies that it receives an empty list
func TestListCertsEmptyList(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/users/11/certificates", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := "{}\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
	}
}

//TestListCertsUser10 requests a list of certificates owned by user 10 and verifies that it receives only the relevant certificates
func TestListCertsUser10(t *testing.T) {

	// First, add the deleted certificate 2
	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/2", bytes.NewBuffer(cert2))
	executeRequest(req)

	// Now add certificate 3, which is owned by a different user
	req, _ = http.NewRequest("POST", "http://localhost:8080/certificates/3", bytes.NewBuffer(cert3))
	executeRequest(req)

	// Now ask for the list of certificates owned by user 10
	req, _ = http.NewRequest("GET", "http://localhost:8080/users/10/certificates", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// Verify that the returned JSON contains only the first two certificates
	expected := `{"1":` + string(cert1Updated) + `,"2":` + string(cert2) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("%v", err)
	}
}

//TestCreateTransfer requests to create a transfer of certificate 1 and then verifies that the transfer has been created
func TestCreateTransfer(t *testing.T) {
	xfer := []byte(`{"to": "test12@test.com","status": "Requested"}`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/1/transfers", bytes.NewBuffer(xfer))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	expected := `{"id":"1","title":"Updated cert","createdAt":"29 MAR 2019","ownerId":"10","year":2019,"note":"This is the updated first certificate","transfer":{"to":"test12@test.com","status":"Requested"}}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("\nError code: %d\n", err)
	}
}

//TestCreateTransferOnExistingTransfer requests to create a transfer of certificate 1, which already has a transfer in place, and then verifies that the request returns an error
func TestCreateTransferOnExistingTransfer(t *testing.T) {
	xfer := []byte(`{"to": "test11@test.com","status": "Requested"}`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/1/transfers", bytes.NewBuffer(xfer))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Certificate 1 is already being transferred to test12@test.com.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestCreateTransferToInvalidUser requests to create a transfer of certificate 1 to a non-existing e-mail, and then verifies it receives an error
func TestCreateTransferToInvalidUser(t *testing.T) {
	xfer := []byte(`{"to": "test100@test.com","status": "Requested"}`)

	req, _ := http.NewRequest("POST", "http://localhost:8080/certificates/2/transfers", bytes.NewBuffer(xfer))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Target test100@test.com isn't valid.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestAcceptNonExistingTransfer requests to accept a transfer that hasn't been created, and then verifies it receives an error
func TestAcceptNonExistingTransfer(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/2/transfers", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "No transfer has been requested for certificate 2.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestAcceptTransferInvalidCert requests to accept a transfer tfor an invalid certificate ID, and then verifies it receives an error
func TestAcceptTransferInvalidCert(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/4/transfers", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	expected := "Certificate ID 4 doesn't exist. Cannot accept transfer.\n"
	if body := response.Body.String(); body != expected {
		t.Errorf("\nExpected %sGot\t %s", expected, body)
	}
}

//TestAcceptTransfer accepts the transfer of certificate 1 and then lists the certificates owned by user 12 to verify that the trtansfer has been completed
func TestAcceptTransfer(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://localhost:8080/certificates/1/transfers", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "http://localhost:8080/users/12/certificates", nil)
	response = executeRequest(req)

	cert1Xferred := []byte(`{"id":"1","title":"Updated cert","createdAt":"29 MAR 2019","ownerId":"12","year":2019,"note":"This is the updated first certificate","transfer":{"to":"","status":""}}`)
	expected := `{"1":` + string(cert1Xferred) + `}`
	body := response.Body.String()
	pass, err := IsEqualJSON(body, expected)

	if !pass {
		t.Errorf("\nExpected %s\nGot\t %s", expected, body)
		t.Errorf("\nError code: %d\n", err)
	}
}

func TestMain(m *testing.M) {
	certificates = make(certsMap) // Initialise the certificates map
	users = make(usersMap)        // Initiatialise the users map

	/* Create some test users data */
	users = make(usersMap) // Initiatialise the users map
	users["10"] = user{"10", "test10@test.com", "Test User 10"}
	users["11"] = user{"11", "test11@test.com", "Test User 11"}
	users["12"] = user{"12", "test12@test.com", "Test User 12"}

	// run tests
	os.Exit(m.Run())
}
