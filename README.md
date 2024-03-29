# Implementation of a chat client using golang

## Description
Simple chat client implementation using golang.
The backend is hosted at https://easybits.tech/ .
Configure an API-controller to point to the running application to the endpoint `/backend`.

## Dependencies

## Usage
Compile the code using the following commands:
```
go build
```

Then run the compiled code:
```
./easy-going -port PORT -easybits_url URL -bearer TOKEN
```
with
- `PORT` the port to run the server on
- `URL` the url of the backend (you get that from the communication channel!)
- `TOKEN` the token to authenticate with the backend (you get that from the communication channel (`api_key`)!)


