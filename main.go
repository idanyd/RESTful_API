// Copyright 2019 Idan Dekel. All rights reserved.

/* This is a RESTful API used to handle certificates creation and update
* You can run it by calling:
* go run main.go
* You can run the unit tests by calling:
* go test -v
*
* The following actions are supported:
* Create a certificate with ID CertID by sending a POST request to [website]/certificates/[CertID] with the following body:
{
    "id": string,
    "title": string,
    "createdAt": string,
    "ownerId": string,
    "year": number,
    "note": string,
    "transfer": {"to":"","status":""}
}
* Update a certificate with ID CertID by sending a PUT request to [website]/certificates/[CertID] with the following body:
{
    "id": (string),
    "title": (string),
    "createdAt": (string),
    "ownerId": (string),
    "year": (number),
    "note": (string),
    "transfer": {"to":"","status":""}
}
* Delete a certificate with ID CertID by sending a DELETE request to [website]/certificates/[CertID] with an empty body
* List all certificates owned by user UserID by sending a GET request to [website]/users/[CertID]/certificates  with an empty body
* Transfer certificate with ID CertID to a different user by sending a POST request to [website]/certificates/[CertID]/transfers with the following body:
{
    "to": [User's e-mail address] (string),
    "status": "Requested"
}
* Accept a transfer of certificate with ID CertID by sending a PUT request to [website]/certificates/[CertID]/transfers  with an empty body
*/

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type transfer struct {
	To     string `json:"to"` /* email address of the recepient */
	Status string `json:"status"`
}

type certificate struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	CreatedAt string   `json:"createdAt"`
	OwnerID   string   `json:"ownerId"`
	Year      int      `json:"year"`
	Note      string   `json:"note"`
	Transfer  transfer `json:"transfer"`
}

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type certsMap map[string]certificate
type usersMap map[string]user

// certificates holds all the existing certificates, mapped by the certificate's Id
var certificates certsMap

// users holds all the currently defined users
var users usersMap

// CreateCert creates a certificate and adds it to the certificates array
func createCert(w http.ResponseWriter, r *http.Request) {
	var cert certificate

	_ = json.NewDecoder(r.Body).Decode(&cert) // Populate cert with the received payload

	if _, ok := (certificates[cert.ID]); ok {
		http.Error(w, "Certificate ID "+cert.ID+" already exists. Cannot create certificate.", http.StatusBadRequest)
	} else if _, ok := users[cert.OwnerID]; !ok {
		http.Error(w, "User ID "+cert.OwnerID+" is invalid. Cannot create certificate.", http.StatusBadRequest)
	} else {
		certificates[cert.ID] = cert            // add the newly-created certificate to the certificates map
		json.NewEncoder(w).Encode(certificates) // Return a JSON with the current certificates
	}
}

// updateCert updates an existing certificate
func updateCert(w http.ResponseWriter, r *http.Request) {
	var cert certificate
	_ = json.NewDecoder(r.Body).Decode(&cert) // Populate cert with the received payload

	if _, ok := (certificates[cert.ID]); !ok {
		http.Error(w, "Certificate ID "+cert.ID+" doesn't exist. Cannot update certificate.", http.StatusBadRequest)
	} else if _, ok := users[cert.OwnerID]; !ok {
		http.Error(w, "User ID "+cert.OwnerID+" is invalid. Cannot update certificate.", http.StatusBadRequest)
	} else {
		certificates[cert.ID] = cert            // add the newly-created certificate to the certificates map
		json.NewEncoder(w).Encode(certificates) // Return a JSON with the current certificates
	}
}

// deleteCert deletes a existing certificate from the certificates map
func deleteCert(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	certID := params["id"]

	if _, ok := (certificates[certID]); !ok {
		http.Error(w, "Certificate ID "+certID+" doesn't exist. Cannot delete certificate.", http.StatusBadRequest)
	} else {
		delete(certificates, certID) // remove the certificate from the certificates map

		var cert certificate
		_ = json.NewDecoder(r.Body).Decode(&cert) // Populate cert with the received payload
		json.NewEncoder(w).Encode(certificates)   // Return a JSON with the current certificates
	}
}

// listCerts lists all certificates held by the user with this id
func listCerts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	if _, ok := (users[userID]); !ok {
		http.Error(w, "User ID "+userID+" is invalid. Cannot list certificates.", http.StatusBadRequest)
	} else {
		// Copy the certificates held by the user from the certificates map into a new map
		certs := make(certsMap)
		for i := range certificates {
			if certificates[i].OwnerID == userID {
				certs[i] = certificates[i]
			}
		}
		json.NewEncoder(w).Encode(certs) // Return a JSON with the user's certificates
	}
}

//createTransfer creates a certificate transfer action
func createTransfer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	certID := params["id"]

	// Workaround that allows us to assign a transfer to an existing certificate in the certificates map
	cert := certificates[certID]
	// Make sure that the certificate is not in the process of being transferred
	if cert.Transfer != (transfer{}) {
		http.Error(w, "Certificate "+certID+" is already being transferred to "+cert.Transfer.To+".", http.StatusBadRequest)
	} else {
		_ = json.NewDecoder(r.Body).Decode(&cert.Transfer)

		targetIsValid := false
		for i := range users {
			if users[i].Email == cert.Transfer.To {
				targetIsValid = true
				break
			}
		}

		if targetIsValid {
			// Update the certificates map only if the target user is valid
			certificates[certID] = cert
			json.NewEncoder(w).Encode(cert) // Return a JSON with the updated certificate
		} else {
			http.Error(w, "Target "+cert.Transfer.To+" isn't valid.", http.StatusBadRequest)
		}
	}
}

//acceptTransfer accepts a trasfer of certificate
func acceptTransfer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	certID := params["id"]

	if _, ok := (certificates[certID]); !ok {
		http.Error(w, "Certificate ID "+certID+" doesn't exist. Cannot accept transfer.", http.StatusBadRequest)
	} else {
		// Workaround that allows us to assign a transfer to an existing certificate in the certificates map
		cert := certificates[certID]

		// Make sure that the transfer request is still active
		if cert.Transfer.Status != "Requested" {
			http.Error(w, "No transfer has been requested for certificate "+certID+".", http.StatusBadRequest)
		} else {
			for i := range users {
				if users[i].Email == cert.Transfer.To {
					// Update the certificate's owner
					cert.OwnerID = users[i].ID
					// Clear the transfer object, as the transfer is complete
					cert.Transfer = (transfer{})
					// Update the global certificates struct
					certificates[certID] = cert
				}
			}
		}
	}
}

// handleRequests handles all HTTP requests
func handleRequests() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/certificates/{id}", createCert).Methods("POST")
	router.HandleFunc("/certificates/{id}", updateCert).Methods("PUT")
	router.HandleFunc("/certificates/{id}", deleteCert).Methods("DELETE")

	router.HandleFunc("/users/{id}/certificates", listCerts).Methods("GET")

	router.HandleFunc("/certificates/{id}/transfers", createTransfer).Methods("POST")
	router.HandleFunc("/certificates/{id}/transfers", acceptTransfer).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	certificates = make(certsMap) // Initialise the certificates map
	handleRequests()
}
