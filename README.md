<!--
parent:
  order: false
-->

# Cosmos SDK

![banner](docs/cosmos-sdk-image.jpg)

This is a fork from the original [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) repo. It contains Fetch.ai specific updates required for the test networks. For this reason it is versioned independantly. Please refer to the [releases](https://github.com/fetchai/cosmos-sdk/releases) section for the compatiblity with upstream versions.

The Cosmos-SDK is a framework for building blockchain applications in Golang.
It is being used to build [`Gaia`](https://github.com/cosmos/gaia), the first implementation of the Cosmos Hub.

**WARNING**: The SDK has mostly stabilized, but we are still making some
breaking changes.

**Note**: Requires [Go 1.13+](https://golang.org/dl/)

## Quick Start

To learn how the SDK works from a high-level perspective, go to the [SDK Intro](./docs/intro/overview.md).

If you want to get started quickly and learn how to build on top of the SDK, please follow the [SDK Application Tutorial](https://tutorials.cosmos.network/nameservice/tutorial/00-intro.html). You can also fork the tutorial's repository to get started building your own Cosmos SDK application.

For more, please go to the [Cosmos SDK Docs](./docs/).

## Cosmos Hub Mainnet

The Cosmos Hub application, `gaia`, has moved to its [own repository](https://github.com/cosmos/gaia). Go there to join the Cosmos Hub mainnet and more.

## Scaffolding

If you are starting a new app or a new module we provide a [scaffolding tool](https://github.com/cosmos/scaffold) to help you get started and speed up development. If you have any questions or find a bug, feel free to open an issue in the repo.

## Disambiguation

This Cosmos-SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. As per our agreement, this disambiguation notice will stay here.
