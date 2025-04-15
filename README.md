# Orchestrator Operator

Go based operator for deploying the Orchestrator.
For more comprehensive information about Orchestrator, please refer to
the [Orchestrator Official Documentation](https://www.rhdhorchestrator.io/).

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

## Contributing to the operator

This project is generated using the operator-sdk. For general knowledge on how to use build the go based operator using
the
operator-sdk tool, please visit: [go-based-operator](https://sdk.operatorframework.io/docs/building-operators/golang/).

#### Fork the Project

To fork the project into a local repository, follow
this [guide](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo)
After making code changes, create a pull request and add reviewers from the maintainer list.

#### Code change rules

1. Ensure to run `make generate manifests` after modifying the `orchestrator_types.go` file.
1. Ensure to run `make manifests` if any changes to
the [kubebuilder makers in the reconciler](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/internal/controller/orchestrator_controller.go#L71)
for rbac changes to be propagated in the `config/rbac` directory.
1. Ensure to run `make bundle` for changes in the config directory to be reflected in the bundle directory.
1. Ensure to add unit tests for any new feature or update. Reference
the [kube package](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/internal/controller/kube).

#### Installation

For local installation, login to an OCP cluster. Currently, the operator has been tested only on OCP clusters.\
Run `make docker-build docker-push` to build and push to your local docker repository.\
Run `make deploy` to install the operator on the OCP cluster.\
Apply a Custom Resource - find sample
CR [here](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/config/samples/_v1alpha3_orchestrator.yaml).

#### Uninstallation

Remove the custom resource and run `make undeploy` to remove the operator resources.

**NOTE**: Run `make help` for more information on all potential make targets

#### Prerequisites for dev tools

- go version v1.22.0+
- operator-sdk v1.38.0+
- docker version 27.03+.
- kubectl version v1.20.0+.
- Access to a OCP cluster v4.14+.

## Upgrading the operator

The mechanism for upgrading the operator currently involves removing the existing operator and its operand resources and
installing the new version.
Follow
this [section in the guide](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/docs/main#cleanup) to
clean up resources and follow
this [section in the guide](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/docs/main#installing-the-orchestrator-go-operator)
to install the latest version.

## License

See the [LICENSE](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/LICENSE) file for details.


