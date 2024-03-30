# Implementation of a chat client using golang

## Description
Simple chat client implementation using golang.
The backend is hosted at https://easybits.tech/ .
Configure an API-controller to point to the running application to the endpoint `/backend`.

The service doesn't store any personal data, unless the user decides to include this into a conversation.
Once the conversation is deleted, the data is removed from the service.
A session cookie is used to store the assign a UID to each user and to identify unique sessions.
Apart from that, the service doesn't store any data.

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

## Development
For development I highly recommend using the `ngrok` tool to temporarily expose your local server to the internet.
This way you can test the chat client with the backend without having to deploy the code to a server.

## Production
If you're planning to deploy this onto an edge-device such as a Raspberry Pi, I recommend using [CloudFlare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) to expose the service to the internet.
It's simple, it's free and it's secure. Unfortunately, it's not open-source.
But most of the time, it's the easiest way to expose a service to the internet and most of you probably already have a CloudFlare account.
