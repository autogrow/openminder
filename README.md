# OpenMinder™

[![](https://godoc.org/autogrow/openminder?status.svg)](http://godoc.org/github.com/autogrow/openminder) [![License: CC BY-NC-SA 4.0](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-sa/4.0/)

Part of Autogrows open initiative, this project is an open source API to be used in conjuction with
the open source OpenMinder hat for the RaspberryPi.  When coupled with the hardware this server will read 
and report via a REST API the readings for:

* 2 EC probes
* 2 pH probes
* 2 tipping buckets

This allows anyone to build a device that can monitor the rootzone of their plants to make the most optimum
use of water and fertigation ingredients to ensure a happy plant.  This is done by measuring the water going
into the plants on the irrigation side, as well as coming out on the runoff side, thus allowing comparisons.

You can also use the [Sample OpenMinder Dashboard](https://github.com/autogrow/openminder-dash) to add a web interface that shows the readings from the REST API.

## Getting Started

The following is a quick getting started overview:

1. Connect the OpenMinder hat to your Raspberry Pi
1. Install the latest version of Raspbian (other OS may work but have not been tested)
1. Build or download the binary, or install the Debian package
1. Run the binary or [service](https://github.com/autogrow/openminder/tree/master/dpkg/lib/systemd/system/openminder.service)
1. Calibrate your probes and tipping buckets
1. Call the readings endpoint for the API `curl http://<ip>:3232/v1/readings`

## Installing

You can install the package from the [Autogrow Debian Repository](https://packagecloud.io/autogrow/public):

    curl -s https://packagecloud.io/install/repositories/autogrow/public/script.deb.sh | sudo bash
    apt-get install openminder

## Building

You can download pre-built binaries from the [releases page](https://github.com/autogrow/openminder/releases)

You can build the binaries using go:

    go get github.com/autogrow/openminder
    GOARCH=arm go build github.com/autogrow/openminder/cmd/openminder  # API and hat interface
    GOARCH=arm go build github.com/autogrow/openminder/cmd/omcli       # command line client

You can build the package using the [ian](https://github.com/penguinpowernz/go-ian) utility:

    cd dpkg
    ian build
    ian pkg

## Usage

You can simply run the API without any arguments to use the defaults for the hat.

    openminder

You can change the API port like so:

    openminder -p 3131

You can also change the tipping bucket GPIO ports (usually only needed for testing) like so:

    openminder -tb1 GPIO17

See `openminder -h` for more information.

### API Endpoints

The following API endpoints are available:

#### GET /v1/readings

This endpoint retrieves all the readings detected from the OpenMinder hat:

    $ curl http://<ip>:3232/v1/readings
    {
        "irrig_adc": 0,
        "irrig_ph_raw": 0,
        "irrig_ph": 0,
        "runoff_adc": 0,
        "runoff_ph_raw": 0,
        "runoff_ph": 0,
        "irrig_tips": 0,
        "irrig_volume": 0,
        "runoff_tips": 0,
        "runoff_volume": 0,
        "irrig_ec": 0,
        "irrig_ec_raw": 0,
        "irrig_ectemp": 0,
        "runoff_ec": 0,
        "runoff_ec_raw": 0,
        "runoff_ectemp": 0,
        "runoff_ratio": 0
    }

**Return codes**:

* 200 - all good

#### PUT /v1/readings/calibrate/:field/:scale/:offset

This endpoint sets the calibration for a specific reading.  The `:field` segment refers to a
reading field from the `/v1/readings` endpoint.  The calibration consists of a scale and an
offset which is applied to the reading: `(reading * scale) + offset`

    $ curl -XPUT http://<ip>:3232/v1/readings/calibrate/irrig_ph/0.345233332/1.2020233

It will return nothing if successful.

**Return codes**:

* 204 - all good
* 400 - forgot to include `:field`, `:scale` or `:offset`
* 500 - failed to save calibration for unknown reason

## Calibration

Some readings need to be calibrated to make any sense.  For instance to turn the tip count into
a volume reading, the minder needs to know how many millilitres each tip of the tipping bucket
represents.  To set the runoff tipping bucket to 5 mls per tip you could do:

    curl -XPUT http://<ip>:3232/v1/readings/calibrate/irrig_tb/5.0/0

### Probes

Calibrating the EC and pH probes required the use of the companion CLI tool `omcli`.  This provides
a series of prompts to help calibrate the probe in place.

## Contributing

