# RESTful_API
This is a RESTful API used to handle certificates creation and update

To clone the repository:

You can run it by calling:
```
go run main.go
```
You can run the unit tests by calling:
```
go test -v
```

The following actions are supported:
Create a certificate with ID CertID by sending a POST request to [website]/certificates/[CertID] with the following body:
```
{
    "id": string,
    "title": string,
    "createdAt": string,
    "ownerId": string,
    "year": number,
    "note": string,
    "transfer": {"to":"","status":""}
}
```
Update a certificate with ID CertID by sending a PUT request to [website]/certificates/[CertID] with the following body:
```
{
    "id": (string),
    "title": (string),
    "createdAt": (string),
    "ownerId": (string),
    "year": (number),
    "note": (string),
    "transfer": {"to":"","status":""}
}
```
Delete a certificate with ID CertID by sending a DELETE request to [website]/certificates/[CertID] with an empty body
List all certificates owned by user UserID by sending a GET request to [website]/users/[CertID]/certificates  with an empty body
Transfer certificate with ID CertID to a different user by sending a POST request to [website]/certificates/[CertID]/transfers with the following body:
```
{
    "to": [User's e-mail address] (string),
    "status": "Requested"
}
```
Accept a transfer of certificate with ID CertID by sending a PUT request to [website]/certificates/[CertID]/transfers  with an empty body
