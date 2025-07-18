# Orchestrator Documentation

For comprehensive documentation on the Orchestrator, please
visit [https://www.rhdhorchestrator.io](https://www.rhdhorchestrator.io).

## Installing the Orchestrator Go Operator

Deploy the Orchestrator solution suite in an OCP cluster using the Orchestrator operator.\
The operator installs the following components onto the target OpenShift cluster:

- RHDH (Red Hat Developer Hub) Backstage
- OpenShift Serverless Logic Operator (with Data-Index and Job Service)
- OpenShift Serverless Operator
    - Knative Eventing
    - Knative Serving
- (Optional) An ArgoCD project named `orchestrator`. Requires an pre-installed ArgoCD/OpenShift GitOps instance in the
  cluster. Disabled by default
- (Optional) Tekton tasks and build pipeline. Requires an pre-installed Tekton/OpenShift Pipelines instance in the
  cluster. Disabled by default

## Important Note for ARM64 Architecture Users

Note that as of November 6, 2023, OpenShift Serverless Operator is based on RHEL 8 images which are not supported on the
ARM64 architecture. Consequently, deployment of this operator on
an [OpenShift Local](https://www.redhat.com/sysadmin/install-openshift-local) cluster on MacBook laptops with M1/M2
chips is not supported.

## Prerequisites

- Logged in to a Red Hat OpenShift Container Platform (version 4.14 +) cluster as a cluster administrator.
- [OpenShift CLI (oc)](https://docs.openshift.com/container-platform/4.13/cli_reference/openshift_cli/getting-started-cli.html)
  is installed.
- [Operator Lifecycle Manager (OLM)](https://olm.operatorframework.io/docs/getting-started/) has been installed in your
  cluster.
- Your cluster has
  a [default storage class](https://docs.openshift.com/container-platform/4.13/storage/container_storage_interface/persistent-storage-csi-sc-manage.html)
  provisioned.
- A GitHub API Token - to import items into the catalog, ensure you have a `GITHUB_TOKEN` with the necessary permissions
  as detailed [here](https://backstage.io/docs/integrations/github/locations/).
    - For classic token, include the following permissions:
        - repo (all)
        - admin:org (read:org)
        - user (read:user, user:email)
        - workflow (all) - required for using the software templates for creating workflows in GitHub
    - For Fine grained token:
        - Repository permissions: **Read** access to metadata, **Read** and **Write** access to actions, actions
          variables, administration, code, codespaces, commit statuses, environments, issues, pull requests, repository
          hooks, secrets, security events, and workflows.
        - Organization permissions: **Read** access to members, **Read** and **Write** access to organization
          administration, organization hooks, organization projects, and organization secrets.

> <font color="red">⚠️**Warning**:</font> Skipping these steps will prevent the Orchestrator from functioning properly.

### Deployment with GitOps

If you plan to deploy in a GitOps environment, make sure you have installed the `ArgoCD/Red Hat OpenShift GitOps` and
the `Tekton/Red Hat Openshift Pipelines Install` operators following
these [instructions](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/gitops/README.md).
The Orchestrator installs RHDH and imports software templates designed for bootstrapping workflow development. These
templates are crafted to ease the development lifecycle, including a Tekton pipeline to build workflow images and
generate workflow K8s custom resources. Furthermore, ArgoCD is utilized to monitor any changes made to the workflow
repository and to automatically trigger the Tekton pipelines as needed.

- `ArgoCD/OpenShift GitOps` operator
    - Ensure at least one instance of `ArgoCD` exists in the designated namespace (referenced by `ARGOCD_NAMESPACE`
      environment variable).
      Example [here](https://raw.githubusercontent.com/rhdhorchestrator/orchestrator-go-operator/main/docs/gitops/resources/argocd-example.yaml)
    - Validated API is `argoproj.io/v1alpha1/AppProject`
- `Tekton/OpenShift Pipelines` operator
    - Validated APIs are `tekton.dev/v1beta1/Task` and `tekton.dev/v1/Pipeline`
    - Requires ArgoCD installed since the manifests are deployed in the same namespace as the ArgoCD instance.

  Remember to
  enable [argocd](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/release-1.6/config/samples/_v1alpha3_orchestrator.yaml#L51)
  in your CR instance.

## Detailed Installation Guide

### From OperatorHub

1. Deploying PostgreSQL reference implementation
    - **If you do not have a PostgreSQL instance in your cluster** \
      you can deploy the PostgreSQL reference implementation by following the
      steps [here](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/postgresql/README.md).
    - **If you already have PostgreSQL running in your cluster** \
      ensure that the default settings in
      the [PostgreSQL values](https://github.com/rhdhorchestrator/orchestrator-helm-chart/blob/main/postgresql/values.yaml)
      file match the `postgres` field provided in
      the [Orchestrator CR](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/release-1.6/config/samples/_v1alpha3_orchestrator.yaml#L23)
      file.
1. Install Orchestrator operator
    1. Go to OperatorHub in your OpenShift Console.
    1. Search for and install the Orchestrator Operator.
1. Run the Setup Script
    1. Follow the steps in the [Running the Setup Script](#running-the-setup-script) section to download and execute the
       setup.sh script, which initializes the RHDH environment.
1. Create an Orchestrator instance
    1. Once the Orchestrator Operator is installed, navigate to Installed Operators.
    1. Select Orchestrator Operator.
    1. Click on Create Instance to deploy an Orchestrator instance.
1. Verify resources and wait until they are running
    1. From console run the following command get the necessary wait commands: \
       `oc describe orchestrator orchestrator-sample -n openshift-operators | grep -A 10 "Run the following commands to wait until the services are ready:"`\

       The command will return an output similar to the one below, which lists several oc wait commands. This depends on
       your specific cluster.
       ```bash
         oc wait -n openshift-serverless deploy/knative-openshift --for=condition=Available --timeout=5m
         oc wait -n knative-eventing knativeeventing/knative-eventing --for=condition=Ready --timeout=5m
         oc wait -n knative-serving knativeserving/knative-serving --for=condition=Ready --timeout=5m
         oc wait -n openshift-serverless-logic deploy/logic-operator-rhel8-controller-manager --for=condition=Available --timeout=5m
         oc wait -n sonataflow-infra sonataflowplatform/sonataflow-platform --for=condition=Succeed --timeout=5m
         oc wait -n sonataflow-infra deploy/sonataflow-platform-data-index-service --for=condition=Available --timeout=5m
         oc wait -n sonataflow-infra deploy/sonataflow-platform-jobs-service --for=condition=Available --timeout=5m
         oc get networkpolicy -n sonataflow-infra
         ```
    1. Copy and execute each command from the output in your terminal. These commands ensure that all necessary services
       and resources in your OpenShift environment are available and running correctly.
    1. If any service does not become available, verify the logs for that service or
       consult [troubleshooting steps](https://www.rhdhorchestrator.io/main/docs/serverless-workflows/troubleshooting/).

### Manual Installation

1. Deploy the PostgreSQL reference implementation for persistence support in SonataFlow following
   these [instructions](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/release-1.6/docs/postgresql/README.md)

1. Create a namespace for the Orchestrator solution:

   ```console
   oc new-project orchestrator
   ```

1. Run the Setup Script
    1. Follow the steps in the [Running the Setup Script](#running-the-setup-script) section to download and execute the
       setup.sh script, which initializes the RHDH environment.

1. Use the following manifest to install the operator in an OCP cluster:

   ```yaml
   apiVersion: operators.coreos.com/v1alpha1
   kind: Subscription
   metadata:
     name: orchestrator-operator
     namespace: openshift-operators
   spec:
     channel: stable
     installPlanApproval: Automatic
     name: orchestrator-operator
     source: redhat-operators
     sourceNamespace: openshift-marketplace
   ```

1. Run the following commands to determine when the installation is completed:

   ```console
   wget https://raw.githubusercontent.com/rhdhorchestrator/orchestrator-go-operator/release-1.6/hack/wait_for_operator_installed.sh -O /tmp/wait_for_operator_installed.sh && chmod u+x /tmp/wait_for_operator_installed.sh && /tmp/wait_for_operator_installed.sh
   ```

   During the installation process, the Orchestrator Operator creates the sub-components
   operators: RHDH operator, OpenShift Serverless operator and OpenShift Serverless Logic operator. Furthermore, it
   creates the necessary CRs and resources needed for orchestrator to function properly.

1. Apply the Orchestrator custom resource (CR) on the cluster to create an instance of RHDH and resources of
   OpenShift
   Serverless Operator and OpenShift Serverless Logic Operator.
   Make any changes to
   the [CR](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/release-1.6/config/samples/_v1alpha3_orchestrator.yaml)
   before applying it, or test the default Orchestrator CR:
    ```console
    oc apply -n orchestrator -f https://raw.githubusercontent.com/rhdhorchestrator/orchestrator-go-operator/refs/heads/release-1.6/config/samples/_v1alpha3_orchestrator.yaml
    ```
   Note: After the first reconciliation of the Orchestrator CR, changes to some of the fields in the CR may not be
   propagated/reconciled to the intended resource. For example, changing the `platform.resources.requests` field in
   the Orchestrator CR will not have any effect on the running instance of the SonataFlowPlatform (SFP) resource.
   For the sake of simplicity, that is the current design and may be revisited in the near future. Please refer to
   the [CRD Parameter List](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/release-1.6/docs/crd)
   to know which fields can be reconciled.

### Running The Setup Script

The setup.sh script simplifies the initialization of the RHDH environment by creating the required authentication secret
and labeling GitOps namespaces based on the cluster configuration.

1. Create a namespace for the RHDH instance. This namespace is predefined as the default in both the setup.sh script and
   the Orchestrator CR but can be overridden if needed.

   ```console
   oc new-project rhdh
   ```

1. Download the setup script from the github repository and run it to create the RHDH secret and label the GitOps
   namespaces:

   ```console
   wget https://raw.githubusercontent.com/rhdhorchestrator/orchestrator-go-operator/release-1.6/hack/setup.sh -O /tmp/setup.sh && chmod u+x /tmp/setup.sh
   ```

1. Run the script:
   ```console
   /tmp/setup.sh --use-default
   ```

**NOTE:** If you don't want to use the default values, omit the `--use-default` and the script will prompt you for
input.

The contents will vary depending on the configuration in the cluster. The following list details all the keys that can
appear in the secret:

> - `BACKEND_SECRET`: Value is randomly generated at script execution. This is the only mandatory key required to be in
    the secret for the RHDH Operator to start.
> - `K8S_CLUSTER_URL`: The URL of the Kubernetes cluster is obtained dynamically using `oc whoami --show-server`.
> - `K8S_CLUSTER_TOKEN`: The value is obtained dynamically based on the provided namespace and service account.
> - `GITHUB_TOKEN`: This value is prompted from the user during script execution and is not predefined.
> - `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`: The value for both these fields are used to authenticate against
    GitHub. For more information open this [link](https://backstage.io/docs/auth/github/provider/).
> - `GITLAB_HOST` and `GITLAB_TOKEN`: The value for both these fields are used to authenticate against
    GitLab.
> - `ARGOCD_URL`: This value is dynamically obtained based on the first ArgoCD instance available.
> - `ARGOCD_USERNAME`: Default value is set to `admin`.
> - `ARGOCD_PASSWORD`: This value is dynamically obtained based on the ArgoCD instance available.

Keys will not be added to the secret if they have no values associated. So for instance, when deploying in a cluster
without the GitOps operators, the `ARGOCD_URL`, `ARGOCD_USERNAME` and `ARGOCD_PASSWORD` keys will be omitted in the
secret.

Sample of a secret created in a GitOps environment:

```console
$> oc get secret -n rhdh -o yaml backstage-backend-auth-secret
apiVersion: v1
data:
  ARGOCD_PASSWORD: ...
  ARGOCD_URL: ...
  ARGOCD_USERNAME: ...
  BACKEND_SECRET: ...
  GITHUB_TOKEN: ...
  K8S_CLUSTER_TOKEN: ...
  K8S_CLUSTER_URL: ...
kind: Secret
metadata:
  creationTimestamp: "2024-05-07T22:22:59Z"
  name: backstage-backend-auth-secret
  namespace: rhdh-operator
  resourceVersion: "4402773"
  uid: 2042e741-346e-4f0e-9d15-1b5492bb9916
type: Opaque
```

### Enabling Monitoring for Workflows

If you want to enable monitoring for workflows, you shall enable it in the `Orchestrator` CR as follows:

```yaml
apiVersion: rhdh.redhat.com/v1alpha3
kind: Orchestrator
metadata:
  name: ...
spec:
  ...
  platform:
    ...
    monitoring:
      enabled: true
      ...
```

After the CR is deployed, follow
the [instructions](https://sonataflow.org/serverlessworkflow/main/cloud/operator/monitoring-workflows.html) to deploy
Prometheus, Grafana and the sample Grafana dashboard.

### Using Knative eventing communication

To enable eventing communication between the different components (Data Index, Job Service and
Workflows), a broker should be used. Kafka is a good candidate as it fulfills the reliability need. You can find the
list of available brokers for Knative is
here: https://knative.dev/docs/eventing/brokers/broker-types/

Alternatively, an in-memory broker could also be used, however it is not recommended to use it for production purposes.

Follow
these [instructions](https://raw.githubusercontent.com/rhdhorchestrator/orchestrator-go-operator/refs/heads/release-1.6/docs/release-1.6/eventing-communication/README.md)
to setup the Knative broker communication.

## Additional information

### Proxy configuration

Your Backstage instance might be configured to work with a proxy. In that case you need to tell Backstage to bypass the
workflow for requests to workflow namespaces and sonataflow namespace (`sonataflow-infra`). You need to add the
namespaces to the environment variable `NO_PROXY`. E.g. NO_PROXY=current-value-of-no-proxy, `.sonataflow-infra`,
`.my-workflow-namespace`. Note the `.` before the namespace name.

### Additional Workflow Namespaces

When deploying a workflow in a namespace different from where Sonataflow services are running (e.g., sonataflow-infra),
several essential steps must be followed:

1. **Allow Traffic from the Workflow Namespace:**
   To allow Sonataflow services to accept traffic from workflows, either create an additional network policy or update
   the existing policy with the new workflow namespace.

   ###### Create Additional Network Policy
   ```console
   oc create -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-external-workflows-to-sonataflow-infra
     # Namespace where network policies are deployed
     namespace: sonataflow-infra
   spec:
     podSelector: {}
     ingress:
       - from:
         - namespaceSelector:
             matchLabels:
               # Allow Sonataflow services to communicate with new/additional workflow namespace.
               kubernetes.io/metadata.name: <new-workflow-namespace>
   EOF
   ```
   ###### Alternatively - Update Existing Network Policy
    ```console
   oc -n sonataflow-infra patch networkpolicy allow-rhdh-to-sonataflow-and-workflows --type='json' \
   -p='[
    {
      "op": "add",
      "path": "/spec/ingress/0/from/-",
      "value": {
        "namespaceSelector": {
          "matchLabels": {
            "kubernetes.io/metadata.name": <new-workflow-namespace>
          }
        }          
      }
    }]'

2. **Identify the RHDH Namespace:**
   Retrieve the namespace where RHDH is running by executing:
   ```console
   oc get backstage -A
   ```
   Store the namespace value in `$RHDH_NAMESPACE` in the Network Policy manifest below.

3. **Identify the Sonataflow Services Namespace:**
   Check the namespace where Sonataflow services are deployed:
   ```console
   oc get sonataflowclusterplatform -A
   ```
   If there is no cluster platform, check for a namespace-specific platform:
   ```console
   oc get sonataflowplatform -A
   ```
   Store the namespace value in `$WORKFLOW_NAMESPACE`.

4. **Set Up a Network Policy:**
   Configure a network policy to allow traffic only between RHDH, Knative, Sonataflow services, and workflows.
   ```console
   oc create -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-rhdh-to-sonataflow-and-workflows
     namespace: $ADDITIONAL_NAMESPACE
   spec:
     podSelector: {}
     ingress:
       - from:
         - namespaceSelector:
             matchLabels:
               # Allows traffic from pods in the RHDH namespace.
               kubernetes.io/metadata.name: $RHDH_NAMESPACE
         - namespaceSelector:
             matchLabels:
               # Allow traffic from pods in the in the Workflow namespace.
               kubernetes.io/metadata.name: $WORKFLOW_NAMESPACE
         - namespaceSelector:
             matchLabels:
               # Allows traffic from pods in the K-Native Eventing namespace.
               kubernetes.io/metadata.name: knative-eventing
         - namespaceSelector:
             matchLabels:
               # Allows traffic from pods in the K-Native Serving namespace.
               kubernetes.io/metadata.name: knative-serving
   EOF
   ```
   To allow unrestricted communication between all pods within the workflow's namespace, create the
   `allow-intra-namespace` network policy.
    ```console
   oc create -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: allow-intra-namespace
     namespace:  $ADDITIONAL_NAMESPACE
   spec:
     # Apply this policy to all pods in the namespace
     podSelector: {}
     # Specify policy type as 'Ingress' to control incoming traffic rules
     policyTypes:
       - Ingress
     ingress:
       - from:
           # Allow ingress from any pod within the same namespace
           - podSelector: {}
   EOF
   ```

5. **Ensure Persistence for the Workflow:**
   If persistence is required, follow these steps:

* **Create a PostgreSQL Secret:**
  The workflow needs its own schema in PostgreSQL. Create a secret containing the PostgreSQL credentials in the
  workflow's namespace:
  ```
  oc get secret sonataflow-psql-postgresql -n sonataflow-infra -o yaml > secret.yaml
  sed -i '/namespace: sonataflow-infra/d' secret.yaml
  oc apply -f secret.yaml -n $ADDITIONAL_NAMESPACE
  ```
* **Configure the Namespace Attribute:**
  Add the namespace attribute under the `serviceRef` property where the PostgreSQL server is deployed.
  ```yaml
  apiVersion: sonataflow.org/v1alpha08
  kind: SonataFlow
    ...
  spec:
    ...
    persistence:
      postgresql:
        secretRef:
          name: sonataflow-psql-postgresql
          passwordKey: postgres-password
          userKey: postgres-username
        serviceRef:
          databaseName: sonataflow
          databaseSchema: greeting
          name: sonataflow-psql-postgresql
          namespace: $POSTGRESQL_NAMESPACE
          port: 5432
  ```
  Replace POSTGRESQL_NAMESPACE with the namespace where the PostgreSQL server is deployed.

By following these steps, the workflow will have the necessary credentials to access PostgreSQL and will correctly
reference the service in a different namespace.

### GitOps environment

See the
dedicated [document](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/gitops/README.md)

### Deploying PostgreSQL reference implementation

See [here](https://github.com/rhdhorchestrator/orchestrator-go-operator/blob/main/docs/postgresql/README.md)

### ArgoCD and workflow namespace

If you manually created the workflow namespaces (e.g., `$WORKFLOW_NAMESPACE`), run this command to add the required
label that allows ArgoCD to deploy instances there:

```console
oc label ns $WORKFLOW_NAMESPACE argocd.argoproj.io/managed-by=$ARGOCD_NAMESPACE
```

### Workflow installation

Follow [Workflows Installation](https://www.rhdhorchestrator.io/serverless-workflows-config/)

## Cleanup

**\/!\\ Before removing the orchestrator, make sure you have first removed any installed workflows. Otherwise the
deletion may become hung in a terminating state.**

To remove the operator, first remove the operand resources

Run:

```console
oc delete namespace orchestrator
```

to delete the Orchestrator CR. This should remove the OSL, Serverless and RHDH Operators, Sonataflow CRs.

To clean up the rest of resources run

```console
oc delete namespace sonataflow-infra rhdh
```

If you want to remove *knative* related resources, you may also run:

```console
oc get crd -o name | grep -e knative | xargs oc delete
```

To remove the operator from the cluster, delete the subscription:

```console
oc delete subscriptions.operators.coreos.com orchestrator-operator -n openshift-operators
```

Note that the CRDs created during the installation process will remain in the cluster.

**Compatibility Matrix between Orchestrator Operator and Dependencies**

| Orchestrator Operator | RHDH  | OSL    | Serverless |
|-----------------------|-------|--------|------------|
| Orchestrator 1.6.0    | 1.6.0 | 1.36.0 | 1.36.0     |

**Compatibility Matrix for Orchestrator Plugins**

| Orchestrator Plugins  Version                                                                           | Orchestrator Operator  Version |
|---------------------------------------------------------------------------------------------------------|--------------------------------|
| Orchestrator Backend (backstage-plugin-orchestrator-backend-dynamic@1.6.0)                              | 1.6.0                          |
| Orchestrator (backstage-plugin-orchestrator@1.6.0)                                                      | 1.6.0                          |
| Orchestrator Scaffolder Backend (backstage-plugin-scaffolder-backend-module-orchestrator-dynamic@1.6.1) | 1.6.0                          |
