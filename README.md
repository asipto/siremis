# SIREMIS - Web Management Interface Fo Kamailio

This is the next-generation SIREMIS written in Go language.

The old SIREMIS generation written in PHP (with support up to PHP 7.x) can be found at:

* https://github.com/asipto/siremis-php

Project web site:

* [https://www.siremis.org](https://www.siremis.org)

## Overview

SIREMIS is a generic web management interface designed to work mainly with Kamailio SIP Server:

* [https://www.kamailio.org](https://www.kamailio.org)

Some screenshots can be seen at:

* [https://github.com/asipto/siremis/issues/1](https://github.com/asipto/siremis/issues/1)

## Usage

Clone the repository:

```
git clone https://github.com/asipto/siremis
```

Build the application:

```
cd siremis
go build .
```

Copy `etc/config-sample.json` to `etc/config.json`. Edit `etc/config.json` and
update database access and authentication users.

Copy `etc/siremis-menu-sample.json` to `etc/siremis-menu.json`, and
`etc/siremis-charts-sample.json` to `etc/siremis-charts.json`.

Run the application:

```
./siremis
```

Go with a modern web browser to:

* http://local.ip:8284/w/

See `siremis -h` for options to set the IP and port to listen on, or the domain
and certificates for HTTPS.

## Version Policy

The version string is composed of three numbers, the format is:

```
YY.MM.VV
```

Where:

* `YY` - the last two digits of the year for the release
* `MM` - the month of the release, if it is `0`, then it is a development version
* `VV` - the incremental version for minor updates of the same release series

## Contributions

Contributions can be made by submitting pull requests and have to be provided
under BSD license.

## License

License type: AGP v3.0

Copyright: 2025 Asipto.com
