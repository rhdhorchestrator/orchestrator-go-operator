# Orchestrator Operator

Go based operator for deploying the Orchestrator.

## Installing the operator

Please visit the [README.md](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/README.md)
page and follow the guide to install the operator in your cluster.

## Releasing the operator

Please visit
the [README.md](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/operator-release/operator-release.md)
page and follow the guide to release the operator.

## Updating the Orchestrator Plugins

Please visit the section in
this [Guide](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/operator-release/operator-release.md#update-the-orchestrator-plugin-if-needed)
and follow the instructions.

## Developing the operator

### Prerequisites

- go version v1.22.0+
- operator-sdk v1.38.0+
- docker version 27.03+.
- kubectl version v1.20.0+.
- Access to a OCP cluster v4.14+.

## Upgrading the operator

The mechanism for upgrading the operator currently involves removing the existing operator and its operand resources and
installing the new version.
Follow
his [section in the guide](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/docs/main#cleanup) to
clean up resources and follow
this [section in the guide](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/docs/main#installing-the-orchestrator-go-operator)
to install the latest version.


