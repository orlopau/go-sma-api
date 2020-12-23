# GoEnergy-API

![build](https://github.com/orlopau/go-sma-api/workflows/build/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/orlopau/go-sma-api/badge.svg?branch=master)](https://coveralls.io/github/orlopau/go-sma-api?branch=master)

GoSMAApi provides tooling for interacting with plants of SMA devices. Included are a CLI Application and a Server
providing an API for interacting with SunSpec compatible plants.

## Table of Contents

* [Energy-CLI](#energy-cli)
    + [Fetch](#fetch)
* [Energy-API](#energy-api)
    + [Configuration](#configuration)
    + [Endpoints](#endpoints)
    + [Docker](#docker)

## Energy-CLI

The application provides a simple command-line application for retrieving data from a plant using SunSpec compatible
devices.

Documentation is provided by calling `energy-cli help`.

### Fetch

`fetch` fetches data from a plant. It automatically detects the device type (e.g. PV-Inverter, Battery-Inverter). The
data is refreshed every 10 seconds.

If the device implements a register containing a valid device address, the address is set automatically. Else, a modbus
slave id can be provided using the `-id` parameter.

This table contains some slave IDs:

| Manufacturer | Slave ID |
| ------------ | -------- |
| SMA          | 126      |

In case the manufacturer of your inverter isn't listed here, you can probably find some documentation by searching for
`< your device > modbus api`.

**Example:**

Given three devices with the following IP addresses:

| Device | IP |
| --- | --- |
| SMA Sunny Boy 2.5 | 192.1.1.1 |
| SMA Sunny Boy 6.0 | 192.1.1.2 |
| SMA Sunny Boy Storage 6.0 | 192.1.1.3 |

First, the modbus (TCP) functionality must be unlocked using the configuration interface of the device. With SMA devices
we can also set a custom port there, the standard port is **502**.

The command should then be `energy-cli fetch -id 126 -a 192.1.1.1:502 -a 192.1.1.2:502 -a 192.1.1.3:502`.

If everything is successful, the output resembles something similar to this:

```
2020/12/21 20:28:54 connecting to 192.1.1.1:502...
2020/12/21 20:28:54 connected to 192.1.1.1:502
2020/12/21 20:28:54 connecting to 192.1.1.2:502...
2020/12/21 20:28:54 connected to 192.1.1.2:502
2020/12/21 20:28:54 connecting to 192.1.1.3:502...
2020/12/21 20:28:54 connected to 192.1.1.3:502
+--------------------+-------+-----+
|      ADDRESS       | POWER | SOC |
+--------------------+-------+-----+
| 192.168.1.1:502    | 123W  |     |
| 192.168.1.2:502    | 300W  |     |
| 192.168.1.3:502    | -20W  | 68% |
+--------------------+-------+-----+
```

## Energy-API

The server provides access to aggregated data of a plant via an HTTP API. A configuration file describing the plant and
its devices is required.

Data is fetched after a message from the plant's energy meter is received, to retrieve the most accurate data for a
point in time. By default, the energymeter sends a message each second, resulting in a refresh interval of 1 second on
the server.

### Configuration

*Environment Variables:*

| Environment Variable | Info | Default |
| --- | --- | --- |
| ENERGY_PORT | Port of the server | 8080 |
| ENERGY_CONFIG_PATH | Path to the plant config | . |

*Plant config:*

The server must be configured using a .yml file, called `plants.yml`.

Example:

```yaml
plant1: # name of the plant
  sunspec: # SunSpec compatible devices
    - "192.168.188.30:502" # ip address pointing to a modbus TCP endpoint
    - "192.168.188.31:502"
    - "192.168.188.32:502"
    - "192.168.188.34:502"
  energymeter: "1901401956" # energymeter serial number
plant2:
  sunspec:
    - "192.168.188.35:502"
    - "192.168.188.36:502"
  energymeter: "3006138525"
```

### Endpoints

`GET /v1/summary` Returns a summary of the energy flow in one or multiple plants. The unit of each value is **watts**.

```json
{
    "plant1": {
        "grid": 1.9,
        "pv": 0,
        "battery": 120,
        "selfConsumption": 121.9,
        "batterySoC": 45,
        "timestampStart": 1608579392,
        "timestampEnd": 1608579392
    },
    "plant2": {
        "grid": 929.2,
        "pv": 0,
        "battery": 0,
        "selfConsumption": 929.2,
        "batterySoC": 14,
        "timestampStart": 1608579392,
        "timestampEnd": 1608579392
    }
}
```

### Docker

A docker image is provided for your convenience. It can be
found [here](https://hub.docker.com/repository/docker/orlopau/go-energy-api).

**Example Usage:**

`docker run -v /path/to/plant.yml:/go/src/app/plants.yml -p 8080:8080 orlopau/go-energy-api`

## Links

* SunSpec modbus resources: https://sunspec.org/sunspec-modbus-home/
* SMA developer documentation: https://www.sma.de/en/products/sma-developer.html

## Licensing

The code in this project is licensed under MIT License.

