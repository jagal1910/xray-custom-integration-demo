# README is a WIP

XRay Custom Integration Demo
-----

### What is a custom integration?

XRay can integrate with external services that provide information about vulnerabilities in packages. If a user wants to implement their own such service they can use a custom integration.

### What this guide covers

- Creating an service that responds to XRay with information about vulnerabilities.

- Integrating the service with XRay.
- Testing the integration on a package with vulnerabilities.
- Testing the integration on a package without vulnerabilities.

### Prerequisites

* go programming language
* [ngrok](https://ngrok.com/download)
* Access to both an XRay instance and an Artifactory instance


### Creating the Service

The service needs to implement only two endpoints. One to **validate an API Key** and another to respond to XRay's requests for info about **package vulnerability**.

## Running the included demo server

`go run main.go (<api-key>) [<path-to-db-file>]`

If a path to db file is not specified, [db.json](./db.json) will be used.
