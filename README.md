# Sim-Load
Artificial load generator for micro-services

## Prerequisites
1. Go (1.9)

## Usage
The service requires that you provided a path to a TOML file following this pattern:
```toml
# Service load can be "light" or "heavy"

[[service]]
    name = "my-service"
    routes = ["/test", "/", "/users"]
    location = "http://localhost:9090"
    load = "light"
[[service]]
    name = "my-service-2"
    routes = ["/"]
    location = "http://localhost:8080"
    load = "heavy"
```

### Running
To run the service on your local machine:
```bash
# build binary
go build

# to see what the flags are:
./sim-load -h

# actually run the program:
./sim-load -services=./services.toml # you can omit the services flag if the toml file is located in the same dir and called "services.toml"
```
If you wish to deploy this (assuming your env is linux), follow these steps:
```bash
# this will create a binary suitable for linux, and a new folder called bin
./build.sh

# copy over the services.toml template
cp services.toml bin

# to deploy just SCP onto your box
scp -r bin user@box:~
```
