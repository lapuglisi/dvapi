# DVApi - a simple REST API written in Go

# Usage
- Fetch this repository to a directory of choice and then try one of the methods below.

## Using Docker
- In the directory you fetched this repo, run:
```bash
$ docker build --tag dvapi-test-api:latest .
```

- Then, run:
```bash
$ docker run --interactive --tty --publish 9098:9098 dvapi-test-api:latest
```
You can also use environment values to customize the API URI. For example:
```bash
$ docker run --interactive --tty \
  --env DVAPI_PORT=8888 --env DVAPI_HOST="127.0.0.1" \
  --publish %{DVAPI_PORT}:%{DVAPI_PORT} dvapi-test-api:latest
```
Making sure that you publish the corresponding port `(%{DVAPI_PORT})` in your `docker run` command.

## Using your linux shell
- First, make sure you have the go binary and its dependencies installed on your environment.
- Then, in the directory you fetched this repo, run the dvapi app.
```bash
$ go run . [-port listen_port] [-host listen_host]

where:
  listen_host: a valid IP address or a valid hostname on which the API will be avaiable (default: 0.0.0.0)
  listen_port: any valid tcp port on which the API will listen (default: 9098)  
```

## Testing the application (with `go test`)
- Just run the following command on the cloned repository root directory:
```bash
$ go test -v .
```

## Consuming the API endpoints
This API implements some endpoints to manage simple devices, as follows:
- ### Creating devices
```bash
curl --request POST ${API_URL}/devices \
--header "Content-Type: application/x-www-form-urlencoded" \
--data '{"name": "device-name", "brand" :"device-brand", "state": "device-state"}'
```
should return
```json
{
  "status": "success",
  "reason": "{'id': new_device_id, 'name': 'device-name', 'brand': 'device-brand', 'state': 'device-state', 'created_on': 'YYYY-mm-ddTHH:MM:SS.?????Z'}"
}
```

- ### Updating devices
```bash
curl --request PATCH ${API_URL}/devices \
--header "Content-Type: application/x-www-form-urlencoded" \
--data '{"id": device_id, "name": "device-name", "brand" :"device-brand", "state": "device-state"}'
```
should return
```json
{
  "status": "success",
  "reason": "device updated succesfully"
}
```
if the device is not in 'in-use' state. Or:
```json
{
  "status": "success",
  "reason": "update device: cannot update a device that is in 'in-use' state"
}
```
if the device is in 'in-use' state.

- ### Deleting devices
```bash
curl --request DELETE ${API_URL}/devices \
--header "Content-Type: application/x-www-form-urlencoded" \
--data '{"id": device_id}'
```
should return:
```json
{
  "status": "success",
  "reason": "device deleted succesfully"
}
```
if the device is not in 'in-use' state. Or:
```json
{
  "status": "success",
  "reason": "delete device: cannot delete a device that is in 'in-use' state"
}
```
if the device is in 'in-use' state.

- ### Fetching all devices
```bash
curl --REQUEST GET ${API_URL}/fetch
```
should return a json array with all devices:
```json
[
  {
    "id": "id",
    "name": "device-name",
    "brand": "device-brand",
    "state": "device-state",
    "created_on": "YYYY-mm-ddTHH:MM:SS.????Z"
  },  
]
```

- ### Fetching a device by id
```bash
curl --REQUEST GET ${API_URL}/fetch/id/{device_id}
```
should return the device json:
```json
[
  {
    "id": "device-id",
    "name": "device-name",
    "brand": "device-brand",
    "state": "device-state",
    "created_on": "YYYY-mm-ddTHH:MM:SS.????Z"
  }  
]
```

- ### Fetching devices by brand(s)
```bash
curl --REQUEST GET ${API_URL}/fetch/brand/{brands_list}

where:
  {brands_list} is a comma delimited string of brands, eg: brand1,brand2,...
```
should return an json array of 0 or mores devices:
```json
[
  {
    "id": "device-id",
    "name": "device-name",
    "brand": "device-brand",
    "state": "device-state",
    "created_on": "YYYY-mm-ddTHH:MM:SS.????Z"
  }  
]
```

- ### Fetching devices by state(s)
```bash
curl --REQUEST GET ${API_URL}/fetch/state/{states_list}

where:
  {states_list} is a comma delimited string of valid states, eg: state1,state2,...
  valid states are: 'available', 'inactive' or 'in-use'
```
should return an json array of 0 or mores devices:
```json
[
  {
    "id": "device-id",
    "name": "device-name",
    "brand": "device-brand",
    "state": "device-state",
    "created_on": "YYYY-mm-ddTHH:MM:SS.????Z"
  }  
]
```

## Issues
- Since the API uses DuckDB as it backing database engine, and DuckDB relies heavily on glibc, alpine is not a viable docker image to containerize the API. Alpine uses musl libaries by default and presents some incompatibility with binaries linked with glibc.
- To be able to use Alpine as a docker image it would be necessary to build duckdb sources on the container, which would be too time consuming.
