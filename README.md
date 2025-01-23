# noble-fiattokenfactory

## Overview

This repository contains the code for noble-fiattokenfactory, a application module built on top of the [Cosmos SDK](https://docs.cosmos.network/main/build/building-modules/intro). It leverages the SDK's capabilities to manage Circle-issued stablecoins on the Noble blockchain.

## Codebase Layout

The codebase is organized into several key directories and files:

- `e2e/`: Includes integration tests for the project.
- `proto/`: Contains all protobuffer definitions for messages used to communicate with the app.
- `simapp/`: Exposes a simulated local application that can be used to test out CLI commands.
- `x/`:
  - `blockibc/`: Contains module execution logic related to interchain operations.
  - `fiattokenfactory/`:
    - `client/`: Contains the main entry points for the application.
    - `keeper/`: Contains core module logic for managing fiat tokens.

## Installation

Follow the steps below to set up your repo locally:

- **Install Golang**:
  - Download and install Golang 1.22 from the official [Golang website](https://go.dev/doc/manage-install).
  - Verify the installation by running `go version`.
- **Install [Heighliner](https://github.com/strangelove-ventures/heighliner)**: Heighliner is a tool that streamlines building cosmos chain containers. Install heighliner by running

    ```sh
    go install github.com/strangelove-ventures/heighliner@v1.6.3
    ```

    Or build heighliner from source by first cloning the [repo](https://github.com/strangelove-ventures/heighliner) and running `go build && go install` inside the cloned repo.

    Run `which heighliner` locally to confirm heighliner is successfully installed.

## Getting Started

### Building the project
You can build the module and simulation app locally using

```sh
make build
```

Once the build step is successful, you would be able to run `simd` in your local terminal to access the CLI application.
Try running `simd query` and `simd tx` to view a list of modules and commands you can interact with.

### Run unit tests
To run unit tests, execute:

```sh
make test-unit
```

### Run Integration Tests

Make sure heighliner is already installed based on instructions in the installation section.

Build the Heighliner image with the local simapp:

```sh
make heighliner
```

Run e2e tests by executing

```sh
make test-e2e
```

Alternatively, you can run all tests (unit and e2e) with:

```sh
make test
```
