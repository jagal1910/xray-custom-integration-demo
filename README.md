XRay Custom Integration Demo
=====

## What is a custom integration?

XRay can integrate with external services that provide information about vulnerabilities in packages. If a user wants to implement their own such service they can use a custom integration.



## Creating a custom Integration

There are three pieces to set up:

- Artifactory
- Your custom integration server
- XRay

### Artifactory Setup

Update the settings for a repository in Artifactory to be viewable by XRay.

![rt-xray-integration-checkbox.png](./images/rt-xray-integration-checkbox.png)

### Running the included demo server

`go run main.go (<api-key>) [<path-to-db-file>]`

If a path to db file is not specified, [db.json](./db.json) will be used. Take note of the api key for the next step.

### Using ngrok to expose your server to the internet

The demo server runs on port 8080, so that's the port to expose.

`ngrok http 8080`

Once ngrok is running, take note of the forwarding urls. These will be provided to XRay when configuring the integration.

![ngrok](./images/ngrok-output.png)

### XRay Setup

Open the integrations view from the admin panel in the sidebar.

![xray-integrations-menu-item](./images/xray-integrations-menu-item.png)

Click the + icon to add an integration.

![./images/add-integration-button](./images/add-integration-button.png)

Select custom integration.

![integration-type](./images/integration-type.png)

Configure the integration.

- The base url will be unique to you (e.g. `https://eq8341dc.ngrok.io`).

- Use `/api/componentinfo` and `/api/checkauth` as the endpoont names.

-  Use `custom-integration-demo` as the Vendor.

  ![integration-config](./images/integration-config.png)

Test the connection and api key by clicking the "Test" button. You should get a message saying "API key is valid" in the XRay UI.



On the XRay homepage, sync XRay's database. This will allow XRay to pick up any Artifactory repositories configured to have XRay enabled.

![sync-db](./images/sync-db.png)

When the sync is done, the component count should increase if XRay finds any new components. The new component should also be visible in the righthand panel.

### Running Tests

`go test ./`