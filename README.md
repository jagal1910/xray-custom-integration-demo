XRay Custom Integration Demo
-----

### What is a custom integration?

XRay can integrate with external services that provide information about vulnerabilities in packages. If a user wants to implement their own such service they can use a custom integration.

### Running the included demo server

- Configure an artifact in artifactory to be viewable by XRay

  

`go run main.go (<api-key>) [<path-to-db-file>]`

If a path to db file is not specified, [db.json](./db.json) will be used.



#### Running Tests

`go test ./`
