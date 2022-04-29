# Lustre Metrics Exporter

<!-- TODO: Create an issue for both, if necessary.
[![Go Report Card](https://goreportcard.com/badge/github.com/HewlettPackard/lustre_exporter)](https://goreportcard.com/report/github.com/HewlettPackard/lustre_exporter)
[![Build Status](https://travis-ci.org/HewlettPackard/lustre_exporter.svg?branch=master)](https://travis-ci.org/HewlettPackard/lustre_exporter)
-->

[Prometheus](https://prometheus.io/) exporter for [Lustre](https://www.lustre.org/) metrics.  

Supported Lustre version is: 2.12.

## Getting

Clone the repository.

## Building

### Exporter

For just building the exporter:  
`make build`

Building the exporter with code testing, formatting and linting:  
`make`

Required for build with code linting:  
`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### RPM Package Build

A CentOS7 RPM package can be build by using a 
1. Shell build script in `rpm/build.sh`
2. Docker build container, see section below.

### Docker - Build Container

#### Binary Build Container

A Debian container is based on the offical golang:1.17.5-bullseye container image for just providing the binary.

```shell
# from repo base dir run
docker build  --tag lustre_exporter -f docker/Dockerfile .
docker run -v $PWD:/cpy -it lustre_exporter
```
The binary will be available in `build/lustre_exporter-X.X.X`.

#### RPM Build Container

A CentOS7 container is based on the official CentOS7 container image for providing the exporter in a RPM package.

```shell
# from repo base dir run
docker build -t rpm_dock -f docker/RPM-Dockerfile .
docker run -v $PWD:/rpm -it rpm_dock
```
The RPM package will be available in `build/x86_64/prometheus-lustre-exporter-X.X.X-X.X.el7.x86_64.rpm`.

## Running

`./lustre_exporter <flags>`

### Flags

* collector.ost=disabled/core/extended
* collector.mdt=disabled/core/extended
* collector.mgs=disabled/core/extended
* collector.mds=disabled/core/extended
* collector.client=disabled/core/extended
* collector.generic=disabled/core/extended
* collector.lnet=disabled/core/extended
* collector.health=disabled/core/extended

All above flags default to the value "extended" when no argument is submitted by the user.

Example: `./lustre_exporter --collector.ost=disabled --collector.mdt=core --collector.mgs=extended`

The above example will result in a running instance of the Lustre Exporter with the following statuses:
* collector.ost=disabled
* collector.mdt=core
* collector.mgs=extended
* collector.mds=extended
* collector.client=extended
* collector.generic=extended
* collector.lnet=extended
* collector.health=extended

Flag Option Detailed Description

- disabled - Completely disable all metrics for this portion of a source.
- core - Enable this source, but only for metrics considered to be particularly useful.
- extended - Enable this source and include all metrics that the Lustre Exporter is aware of within it.

## What's exported?

All Lustre procfs and procsys data from all nodes running the Lustre Exporter that we perceive as valuable data is exported or can be added to be exported (we don't have any known major gaps that anyone cares about, so if you see something missing, please file an issue!).

See the issues tab for all known issues.

## Troubleshooting

In the event that you encounter issues with specific metrics (especially on versions of Lustre older than 2.7), please try disabling those specific troublesome metrics using the documented collector flags in the 'disabled' or 'core' state. Users have encountered bugs within Lustre where specific sysfs and procfs files miscommunicate their sizes, causing read calls to fail.

## Contributing

You are welcome to contribute to the project.
Feel free to create an issue, pull request or just start a discussion.
