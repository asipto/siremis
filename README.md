# SIREMIS - Web Management Interface

Project web site:

* https://www.siremis.org

## Overview

SIREMIS is a generic web management interface designed to work mainly with Kamailio SIP Server:

* https://www.kamailio.org

## Usage

Clone the repository:

```
git clone https://github.com/asipto/siremis-go
```

Build the application:

```
cd siremis-go
go build .
```

Copy `etc/config-sample.json` to `etc/config.json`. Edit `etc/config.json` and
update database access and authentication users.

Copy `etc/siremis-menu-sample.json` to `etc/siremis-menu.json`, and
`etc/siremis-charts-sample.json` to `etc/siremis-charts.json`.

Run the application:

```
./siremis-go
```

Go with a modern web browser to:

* http://local.ip:8284/w/


## Contributions

Contributions can be made by submitting pull requests and have to be provided under BSD license.

## License

License type: AGP v3.0

Copyright: 2025 Asipto.com
