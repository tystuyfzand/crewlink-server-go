[![ISC License][license-shield]][license-url] [![Docker Pulls][docker-shield]][docker-url]

<br />
<p align="center">
  <a href="https://github.com/tystuyfzand/crewlink-server-go">
    <img src="https://github.com/tystuyfzand/crewlink-server-go/raw/master/assets/images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">CrewLink Go Server</h3>

  <p align="center">
    Go Voice Relay server for <a href="https://github.com/ottomated/crewlink">CrewLink</a>.
    <br />
    <a href="https://github.com/tystuyfzand/crewlink-server-go/issues">Report Bug</a>
    ·
    <a href="https://github.com/tystuyfzand/crewlink-server-go/issues">Request Feature</a>
  </p>
</p>



<!-- TABLE OF CONTENTS -->
## Table of Contents

* [About the Project](#about-the-project)
* [Docker Quickstart](#docker-quickstart)
    * [Building the Docker Image](#building-the-docker-image)
* [Getting Started](#getting-started)
    * [Prerequisites](#prerequisites)
    * [Installation](#installation)
* [Contributing](#contributing)
* [License](#license)



<!-- ABOUT THE PROJECT -->
## About The Project

This is the relay server for CrewLink, an Among Us proximity voice chat program, written in Go instead of JavaScript. It is aiming to handle many more users at once, as well as potentially offering an all-in-one solution that includes a STUN/TURN server.

This project is under heavy development and SHOULD work just the same as the base server. It will also support the most relevant Pull Requests for future support.

## Features

Every feature in the main server is currently implemented.

Additionally, the following features are implemented ahead of acceptance into the main server/client:

- [Allow server host to decide peer to peer behavior (STUN/TURN support to bypass NAT issues)](https://github.com/ottomated/CrewLink-server/pull/28)

The following features are planned, and I'd welcome pull requests for:

- Automatic TLS via Go's `autocert` package
- [Room-specific settings](https://github.com/ottomated/CrewLink-server/pull/39)

## Environment Variables

Optional environment variables:

- `ADDRESS`: Specifies the server bind address. The Web UI will show whatever address the client actually connects as.
- `NAME`: Specifies the server name. Only used in the `/health` endpoint.
- `TRUSTED_PROXIES`: Specifies trusted proxy addresses for IP and Scheme detection. Default is `10.0.0.0/8,172.16.0.0/12,192.168.0.0/16`, you should change this if you use a public IP (for example, Cloudflare)
- `LOG_REQUESTS`: Enables request logging from the internal http server. Only logs requests to `/`, `/logo.png`, and the initial websocket connection.
- `CERTIFICATE_PATH`: Specifies directory to search for certificates. Currently, certificates must be named `server.(crt|key)` or Let's Encrypt's default `fullchain.pem` and `privkey.pem`.

## Docker Quickstart

Run the server with [Docker](https://docs.docker.com/get-docker/) by running the following command:

```
docker run -d -p 9736:9736 tystuyfzand/crewlink-server-go:latest
```

To change the external port the server uses, change the *first* instance of the port. For example, to use port 8123:

```
docker run -d -p 8123:9736 tystuyfzand/crewlink-server-go:latest
```

### Building the Docker Image

To build your own Docker image, do the following:

1. Clone the repo
```sh
git clone https://github.com/tystuyfzand/crewlink-server-go.git
cd crewlink-server-go
```

2. Run the Docker build command:
```sh
docker build -t tystuyfzand/crewlink-server:build .
```

## Manual Installation

### Prerequisites

This is an example of how to list things you need to use the software and how to install them.
* [Go 1.15](https://golang.org/dl/)


### Installation

1. Clone the repo
```sh
git clone https://github.com/tystuyfzand/crewlink-server-go.git
cd crewlink-server-go
```
2. Build the binary
```sh
go build -o crewlink cmd/main.go
```
3. Run the project
```JS
./crewlink
```
4. Copy your server's IP and port into CrewLink settings. Make sure everyone in your lobby is using the same server.

<!-- CONTRIBUTING -->
## Contributing

Any contributions you make are greatly appreciated.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


## License

Distributed under the ISC License. See `LICENSE` for more information.

[license-shield]: https://img.shields.io/github/license/tystuyfzand/crewlink-server-go.svg?style=flat-square
[license-url]: https://github.com/tystuyfzand/crewlink-server-go/blob/master/LICENSE
[docker-shield]: https://img.shields.io/docker/pulls/tystuyfzand/crewlink-server-go
[docker-url]: https://hub.docker.com/repository/docker/tystuyfzand/crewlink-server-go