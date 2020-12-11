[![GPL-3.0 License][license-shield]][license-url]

<br />
<p align="center">
  <a href="https://github.com/tystuyfzand/crewlink-server-go">
    <img src="assets/images/logo.png" alt="Logo" width="80" height="80">
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

This project is likely UNUSABLE, but will become a fully usable and feature-complete server in time.

## Environment Variables

Optional environment variables:

- `ADDRESS`: Specifies the server domain
- `NAME`: Specifies the server name

## Docker Quickstart

Coming soon!

### Building the Docker Image

Coming soon!

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
go build -o crewlink
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