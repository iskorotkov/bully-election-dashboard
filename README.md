# Dashboard for bully election algorithm

Dashboard for bully election algorithm for distributed systems.

[![CodeQL](https://github.com/iskorotkov/bully-election-dashboard/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/iskorotkov/bully-election-dashboard/actions/workflows/codeql-analysis.yml)

- [Dashboard for bully election algorithm](#dashboard-for-bully-election-algorithm)
  - [Endpoints](#endpoints)
  - [Build](#build)
  - [Deploy](#deploy)
  - [Links](#links)

## Endpoints

1. `/` - for web GUI app.
1. `/api` - for HTTP API.

## Build

```sh
make build # to build locally.
make test # to run all tests.
make build-image # to build Docker image.
```

## Deploy

```sh
make deploy # to deploy in Kubernetes cluster.
```

## Links

- [Bully election implementation](https://github.com/iskorotkov/bully-election)
- [Bully algorithm - Wikipedia](https://en.wikipedia.org/wiki/Bully_algorithm)
