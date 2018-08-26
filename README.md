# Open Virtual Network (OVN) Exporter

Export Open Virtual Network (OVN) data to Prometheus.

## Introduction

This exporter exports metrics from the following OVN components:
* OVN `northd` service
* OVS `vswitchd` service
* `OVN Northbound` database
* `OVN Southbound` database
* `Open_vSwitch` database

## Getting Started

To run it:

```bash
cd $GOPATH/src
mkdir -p github.com/greenpau
cd github.com/greenpau
git clone https://github.com/greenpau/ovn_exporter.git
cd ovn_exporter
make
make qtest
```

## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| `ovn_up` | Is OVN up and health (1) or is it partially or totally down (0). | `system_id` |

For example:

```
ovn_up{system_id="489403bf-68ec-4f5b-acc5-2b2ade099f41"} 1
```

## Flags

```bash
./bin/ovn-exporter --help
```

* __`ovn.timeout`:__ Timeout on requests to OVN SB, OVN NB, and OVS databases.
* __`ovn.poll-interval`:__ The minimum interval (in seconds) between collections from OVN. (default: 30 seconds).
* __`version`:__ Show application version.
* __`web.listen-address`:__ Address to listen on for web interface and telemetry.
* __`web.telemetry-path`:__ Path under which to expose metrics.
