# DVApi - a simple REST API written in Go

# Usage
- Fetch this repository to a directory of choice and then try one of the methods below.

## Using Docker
- In the directory you fetched this repo, run:
```bash
$ docker build --tag dvapi:latest .
```

- Then, run:
```bash
$ docker run --interactive --tty --publish 9098:9098 dvapi:latest
```

## Using your linux shell
- First, make sure you have the go binary and its dependencies installed on your environment.
- Then, in the directory you fetched this repo, run the dvapi app.
```bash
$ go run . [-port listen_port] [-host listen_host]

where:
  listen_host: a valid IP address or a valid hostname on which the API will be avaiable (default: 0.0.0.0)
  listen_port: any valid tcp port on which the API will listen (default: 9098)  
```
