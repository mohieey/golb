# lbgo: Round Robin Load Balancer in Go

## Overview

`lbgo` is a lightweight Round Robin Load Balancer implemented in Go. It is designed to distribute incoming network traffic across multiple backend servers in a circular manner, ensuring a balanced workload.

## Features

- Simple and efficient Round Robin algorithm.
- Support for dynamic backend server registration and deregistration.
- Health checks to monitor the status of backend servers.

## Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/mohieey/lbgo.git

# Navigate to the project directory
cd lbgo

# Build the project
go build .

# Run!
./lbgo
```

### Configuring the load balancer

- A file named configs.yaml should be in the same path with the project binary,
  to provide a file with different name, the run command will be like this
  `./lbgo -configs <filename>`

- configs file example:

```bash
# The port to listen on
port: 8000

# The nodes that will be balanced between
nodes:
  - http://localhost:3000
  - http://localhost:3001
  - http://localhost:3002
```
