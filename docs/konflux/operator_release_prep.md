# Preparation For Konflux Release

## Table of contents

* [Preparation for Konflux Release](#preparation-for-konflux-release)
    * [Create Go Operator Config For New Release](#create-go-operator-config-for-new-release)
    * [Create File Based Catalog (FBC) Config For New OCP Version](#create-file-based-catalog-fbc-config-for-new-ocp-version)
    * [Run Build Manifest Script](#run-build-manifest-script)
    * [Add ReleasePlanAdmission (RPA) For New Release](#add-releaseplanadmission-rpa-for-new-release)
    * [Create Merge Request](#create-merge-request)
    * [Prep Go Orchestrator Repo For Release](#prep-go-orchestrator-repo-for-release)
    * [Update Main branch Via PR](#update-main-branch-via-pr)

Before releasing with Konflux, it is essential to follow this guide to set up the necessary configuration/manifests and
pipeline for the specific release version.
These configurations are added in
the [Konflux release data repo](https://gitlab.cee.redhat.com/releng/konflux-release-data).
To get access to the repo as a CODEOWNER, please reach out to @jordigilh, @masayag,
@gciavarr, or @jubah.

Once you have the right access, create a branch `release-1.x` from main.

#### Create Go Operator Config For New Release

In the newly created branch `release-1.x`:

* Navigate
  to [orchestrator tenant config](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config/cluster/stone-prd-rh01/tenants/orchestrator-releng-tenant/operator?ref_type=heads)
* Add a new config file `operator-1.x.yaml` for new release.  
  This should contain `Application`, `Component`, `ReleasePlan`,
  `ImageRepository`, and `IntegrationTestScenario`.

  Example for 1.5 release `operator-1-5.yaml`:
  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: Application
  metadata:
    name: operator-1-5
    namespace: orchestrator-releng-tenant
  spec:
    displayName: operator (release-1-5)
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: Component
  metadata:
    name: controller-rhel9-operator-1-5
    namespace: orchestrator-releng-tenant
  spec:
    application: operator-1-5
    build-nudges-ref:
      - orchestrator-operator-bundle-1-5
    componentName: controller-rhel9-operator-1-5
    containerImage: quay.io/redhat-user-workloads/orchestrator-releng-tenant/controller-rhel9-operator
    source:
      git:
        dockerfileUrl: Dockerfile
        revision: main
        url: https://github.com/rhdhorchestrator/orchestrator-go-operator.git
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: Component
  metadata:
    name: orchestrator-operator-bundle-1-5
    namespace: orchestrator-releng-tenant
  spec:
    application: operator-1-5
    componentName: orchestrator-operator-bundle-1-5
    containerImage: quay.io/redhat-user-workloads/orchestrator-releng-tenant/orchestrator-operator-bundle
    source:
      git:
        dockerfileUrl: bundle.konflux.Dockerfile
        revision: main
        url: https://github.com/rhdhorchestrator/orchestrator-go-operator.git
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ImageRepository
  metadata:
    annotations:
      image-controller.appstudio.redhat.com/update-component-image: "true"
    name: imagerepository-for-go-operator-1-5-controller-rhel9-operator-1-5
    namespace: orchestrator-releng-tenant
    labels:
      appstudio.redhat.com/application: operator-1-5
      appstudio.redhat.com/component: controller-rhel9-operator-1-5
  spec:
    image:
      visibility: public
      name: orchestrator-releng-tenant/controller-rhel9-operator
    notifications:
      - config:
          url: https://bombino.api.redhat.com/v1/sbom/quay/push
        event: repo_push
        method: webhook
        title: SBOM-event-to-Bombino
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ImageRepository
  metadata:
    annotations:
      image-controller.appstudio.redhat.com/update-component-image: "true"
    name: imagerepository-for-go-operator-1-5-orchestrator-operator-bundle-1-5
    namespace: orchestrator-releng-tenant
    labels:
      appstudio.redhat.com/application: operator-1-5
      appstudio.redhat.com/component: orchestrator-operator-bundle-1-5
  spec:
    image:
      visibility: public
      name: orchestrator-releng-tenant/orchestrator-operator-bundle
    notifications:
      - config:
          url: https://bombino.api.redhat.com/v1/sbom/quay/push
        event: repo_push
        method: webhook
        title: SBOM-event-to-Bombino
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlan
  metadata:
    name: operator-staging-1-5
    labels:
      release.appstudio.openshift.io/auto-release: "false"
      release.appstudio.openshift.io/releasePlanAdmission: operator-staging-1-5
      release.appstudio.openshift.io/standing-attribution: "true"
  spec:
    application: operator-1-5
    target: rhtap-releng-tenant
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlan
  metadata:
    name: operator-prod-1-5
    labels:
      release.appstudio.openshift.io/auto-release: "false"
      release.appstudio.openshift.io/releasePlanAdmission: operator-prod-1-5
      release.appstudio.openshift.io/standing-attribution: "true"
  spec:
    application: operator-1-5
    target: rhtap-releng-tenant
  ---
  apiVersion: appstudio.redhat.com/v1beta2
  kind: IntegrationTestScenario
  metadata:
    name: operator-enterprise-contract-1-5
    namespace: orchestrator-releng-tenant
  spec:
    params:
      - name: POLICY_CONFIGURATION
        value: rhtap-releng-tenant/registry-orchestrator-releng
      - name: SINGLE_COMPONENT
        value: "true"
    application: operator-1-5
    contexts:
      - description: Application testing
        name: application
    resolverRef:
      params:
        - name: url
          value: "https://github.com/konflux-ci/build-definitions"
        - name: revision
          value: main
        - name: pathInRepo
          value: pipelines/enterprise-contract.yaml
      resolver: git
  ```
* Update
  the [kustomization file](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/blob/main/tenants-config/cluster/stone-prd-rh01/tenants/orchestrator-releng-tenant/operator/kustomization.yaml?ref_type=heads#L4).
  Add the new `operator-1-x.yaml` to the resource list.
  Example for 1.5 release `operator-1-5.yaml`:
  ```yaml
  ---
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  resources:
    - helm-operator-1-3.yaml
    - helm-operator-1-4.yaml
    - operator-1-5.yaml
  ```

#### Create File Based Catalog (FBC) Config For New OCP Version

* Navigate
  to [orchestrator tenant config](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config/cluster/stone-prd-rh01/tenants/orchestrator-releng-tenant?ref_type=heads)
* Add a new folder for the new FBC `fbc-v4-17` for referencing the newly supported OCP version.
  Example folder
  for [v4.17](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/blob/main/tenants-config/cluster/stone-prd-rh01/tenants/orchestrator-releng-tenant/fbc-v4-17/kustomization.yaml?ref_type=heads)
* Add a new config file `fbc-v4-17.yaml` for new OCP version.
  This should contain the Application, Component, ReleasePlan, ImageRepository, IntegrationTestScenario.
  Example for `v4.17` OCP version `fbc-v4-17.yaml`
  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: Application
  metadata:
    name: fbc-v4-17
    namespace: orchestrator-releng-tenant
  spec:
    displayName: FBC v4.17
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: Component
  metadata:
    name: fbc-v4-17
    namespace: orchestrator-releng-tenant
  spec:
    application: fbc-v4-17
    componentName: fbc-v4-17
    containerImage: quay.io/redhat-user-workloads/orchestrator-releng-tenant/fbc-v4-17
    source:
      git:
        context: v4.17
        dockerfileUrl: catalog.Dockerfile
        revision: main
        url: https://github.com/rhdhorchestrator/orchestrator-fbc.git
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ImageRepository
  metadata:
    annotations:
      image-controller.appstudio.redhat.com/update-component-image: "true"
    name: imagerepository-for-fbc-v4-17
    namespace: orchestrator-releng-tenant
    labels:
      appstudio.redhat.com/application: fbc-v4-17
      appstudio.redhat.com/component: fbc-v4-17
  spec:
    image:
      name: orchestrator-releng-tenant/fbc-v4-17
      visibility: public
    notifications:
      - config:
          url: https://bombino.api.redhat.com/v1/sbom/quay/push
        event: repo_push
        method: webhook
        title: SBOM-event-to-Bombino
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlan
  metadata:
    labels:
      release.appstudio.openshift.io/auto-release: "false"
      release.appstudio.openshift.io/releasePlanAdmission: orchestrator-fbc-prod-index-v4-15-plus
      release.appstudio.openshift.io/standing-attribution: "true"
    name: fbc-v4-17-release-as-production-fbc
    namespace: orchestrator-releng-tenant
  spec:
    application: fbc-v4-17
    releaseGracePeriodDays: 7
    target: rhtap-releng-tenant
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlan
  metadata:
    name: fbc-v4-17-release-as-staging-fbc
    namespace: orchestrator-releng-tenant
    labels:
      release.appstudio.openshift.io/auto-release: "false"
      release.appstudio.openshift.io/releasePlanAdmission: orchestrator-fbc-staging-index-v4-15-plus
      release.appstudio.openshift.io/standing-attribution: "true"
  spec:
    application: fbc-v4-17
    target: rhtap-releng-tenant
  ---
  apiVersion: appstudio.redhat.com/v1beta2
  kind: IntegrationTestScenario
  metadata:
    name: fbc-v4-17-enterprise-contract
    namespace: orchestrator-releng-tenant
  spec:
    application: fbc-v4-17
    params:
      - name: POLICY_CONFIGURATION
        value: rhtap-releng-tenant/fbc-stage
    resolverRef:
      params:
        - name: url
          value: "https://github.com/konflux-ci/build-definitions"
        - name: revision
          value: main
        - name: pathInRepo
          value: pipelines/enterprise-contract.yaml
      resolver: git
  ```
* Add the `kustomization.yaml` file under the same folder.
  Example for `v4.17`
  ```yaml
  ---
  kind: Kustomization
  apiVersion: kustomize.config.k8s.io/v1beta1
  # Naming: <API_GROUP>/<KIND_PLURAL>/<METADATA_NAME>
  resources:
    - fbc-v4-17.yaml
  ```

#### Run Build Manifest Script

* Run the `build-manifests.sh` script (
  found [here](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config?ref_type=heads)).
  This will add or update the manifests under
  the [auto-generated folder](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/tenants-config/auto-generated/cluster/stone-prd-rh01/tenants/orchestrator-releng-tenant?ref_type=heads).
  Commit these change in addition to any other relevant additions.

#### Add ReleasePlanAdmission (RPA) For New Release

* Navigate
  to [orchestrator RPA config folder](https://gitlab.cee.redhat.com/releng/konflux-release-data/-/tree/main/config/stone-prd-rh01.pg1f.p1/product/ReleasePlanAdmission/orchestrator-releng?ref_type=heads)
* Add a new RPA for the staging Go Operator and follow naming convention `operator-staging-1.x.yaml.`
  Example of staging RPA manifest for 1.5 release:
  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlanAdmission
  metadata:
    labels:
      release.appstudio.openshift.io/auto-release: "true"
      pp.engineering.redhat.com/business-unit: application-developer
    name: operator-staging-1-5
    namespace: rhtap-releng-tenant
  spec:
    applications:
      - operator-1-5
    origin: orchestrator-releng-tenant
    policy: registry-orchestrator-releng
    data:
      releaseNotes:
        product_id: 851
        product_name: RHDH
        product_version: "1.5"
        type: "RHBA"
        synopsis: "Red Hat Developer Hub Orchestrator"
        topic: |
          The developer preview release of Red Hat Developer Hub Orchestrator.
        description: |
          Red Hat Developer Hub Orchestrator is a plugin that enables serverless asynchronous workflows to Backstage.
          This plugin is a development preview release.
        solution: |
          RHDH Orchestrator introduces serverless asynchronous workflows to Backstage, with a focus on facilitating the
          transition of applications to the cloud, onboarding developers, and enabling users to create workflows for
          backstage actions or external systems.
        references:
          - https://www.redhat.com/en/technologies/cloud-computing/developer-hub
          - https://rhdhorchestrator.io
      sign:
        configMapName: "hacbs-signing-pipeline-config-staging-redhatbeta2"
        cosignSecretName: konflux-cosign-signing-stage
      mapping:
        components:
          - name: controller-rhel9-operator-1-5
            repository: "registry.stage.redhat.io/rhdh-orchestrator-dev-preview-beta/controller-rhel9-operator"
          - name: orchestrator-operator-bundle-1-5
            repository: "registry.stage.redhat.io/rhdh-orchestrator-dev-preview-beta/orchestrator-operator-bundle"
        defaults:
          tags:
            - "1.5"
            - "1.5-{{ timestamp }}"
            - "{{ git_sha }}"
            - "{{ git_short_sha }}"
          pushSourceContainer: true
      pyxis:
        secret: pyxis-staging-secret
        server: stage
    pipeline:
      pipelineRef:
        resolver: git
        params:
          - name: url
            value: https://github.com/konflux-ci/release-service-catalog.git
          - name: revision
            value: production
          - name: pathInRepo
            value: "pipelines/managed/rh-advisories/rh-advisories.yaml"
      serviceAccountName: release-registry-staging
      timeouts:
        pipeline: "01h0m0s"
        tasks: 01h0m0s
  ```
* Add a new RPA for the production go operator and follow naming convention `operator-prod-1.x.yaml.`
  Example of production RPA manifest for 1.5 release:
  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlanAdmission
  metadata:
    labels:
      release.appstudio.openshift.io/auto-release: "true"
      pp.engineering.redhat.com/business-unit: application-developer
    name: operator-prod-1-5
    namespace: rhtap-releng-tenant
  spec:
    applications:
      - operator-1-5
    origin: orchestrator-releng-tenant
    policy: registry-orchestrator-releng
    data:
      releaseNotes:
        product_id: 851
        product_name: RHDH
        product_version: "1.5"
        type: "RHBA"
        synopsis: "Red Hat Developer Hub Orchestrator"
        topic: |
          The developer preview release of Red Hat Developer Hub Orchestrator.
        description: |
          Red Hat Developer Hub Orchestrator is a plugin that enables serverless asynchronous workflows to Backstage.
          This plugin is a development preview release.
        solution: |
          RHDH Orchestrator introduces serverless asynchronous workflows to Backstage, with a focus on facilitating the
          transition of applications to the cloud, onboarding developers, and enabling users to create workflows for
          backstage actions or external systems.
        references:
          - https://www.redhat.com/en/technologies/cloud-computing/developer-hub
          - https://rhdhorchestrator.io
      sign:
        configMapName: "hacbs-signing-pipeline-config-redhatbeta2"
        cosignSecretName: konflux-cosign-signing-stage
      mapping:
        components:
          - name: controller-rhel9-operator-1-5
            repository: "registry.redhat.io/rhdh-orchestrator-dev-preview-beta/controller-rhel9-operator"
          - name: orchestrator-operator-bundle-1-5
            repository: "registry.redhat.io/rhdh-orchestrator-dev-preview-beta/orchestrator-operator-bundle"
        defaults:
          tags:
            - "1.5"
            - "1.5-{{ timestamp }}"
          pushSourceContainer: true
      pyxis:
        secret: pyxis-prod-secret
        server: production
    pipeline:
      pipelineRef:
        resolver: git
        params:
          - name: url
            value: "https://github.com/konflux-ci/release-service-catalog.git"
          - name: revision
            value: production
          - name: pathInRepo
            value: "pipelines/managed/rh-advisories/rh-advisories.yaml"
      serviceAccountName: release-registry-prod
      timeouts:
        pipeline: "01h0m0s"
        tasks: 01h0m0s
  ```
* Update the staging RPA for FBC Index.
  If necessary, update the RPA for the FBC index when we want to support a new OCP version.
  Update the existing `orchestrator-fbc-staging-index-v4-15-plus` by adding the new FBC under
  the `spec.applications` list.
  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlanAdmission
  metadata:
    labels:
      release.appstudio.openshift.io/auto-release: "true"
      pp.engineering.redhat.com/business-unit: application-developer
    name: orchestrator-fbc-staging-index-v4-15-plus
    namespace: rhtap-releng-tenant
  spec:
    applications:
      - fbc-v4-15
      - fbc-v4-16
      - fbc-v4-17
    data:
      releaseNotes:
        product_id: 851
        product_name: RHDH
        product_version: fbc
        ...
    origin: orchestrator-releng-tenant
    policy: fbc-stage
  ```
* Update the production RPA for FBC Index.
  If necessary, update the RPA for the FBC index when we want to support a new OCP version.
  Update the existing `orchestrator-fbc-prod-index-v4-15-plus` by adding the new FBC under
  the `spec.applications` list.

  ```yaml
  ---
  apiVersion: appstudio.redhat.com/v1alpha1
  kind: ReleasePlanAdmission
  metadata:
    labels:
      release.appstudio.openshift.io/auto-release: "true"
      pp.engineering.redhat.com/business-unit: application-developer
    name: orchestrator-fbc-prod-index-v4-15-plus
    namespace: rhtap-releng-tenant
  spec:
    applications:
      - fbc-v4-15
      - fbc-v4-16
      - fbc-v4-17
    data:
      releaseNotes:
        product_id: 851
        product_name: RHDH
        product_version: fbc
        ...
    origin: orchestrator-releng-tenant
    policy: registry-orchestrator-fbc-prod-with-weekends
  ```

#### Create Merge Request

* After pushing the changes, create a merge request, have it reviewed and
  approved.
* After the merge, check the Konflux UI to ensure that your configuration
  changes have been applied successfully. This is facilitated by the existing
  _ArgoCD_ setup, which automatically syncs changes from the Git repository to the
  environment.

#### Prep Go Orchestrator Repo For Release

Once development is complete and QE gives the green light, create a branch from main

* Navigate to the [.tekton folder](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/.tekton)
  and update the pipeline files names to suffix with `xxx-1.x.yaml`.  
  Example for 1.5 release branch:
  ```console
  - controller-rhel9-operator-on-pull-request-1-5.yaml
  - controller-rhel9-operator-on-push-1-5.yaml
  - orchestrator-operator-bundle-on-pull-request-1-5.yaml
  - orchestrator-operator-bundle-on-push-1-5.yaml
  ```
* In each of the pipeline config file listed above, in the `pipelinesascode.tekton.dev/on-cel-expression`,
  update the `target_branch` from main to `release-1.x`
* Update the `labels`, `name`, `component` and any relevant change in each pipeline config.

  Example of 1.5 `controller-rhel9-operator-on-pull-request-1-5.yaml`:

  ```yaml
  apiVersion: tekton.dev/v1
  kind: PipelineRun
  metadata:
    annotations:
      build.appstudio.openshift.io/repo: https://github.com/rhdhorchestrator/orchestrator-go-operator?rev={{revision}}
      build.appstudio.redhat.com/commit_sha: '{{revision}}'
      build.appstudio.redhat.com/pull_request_number: '{{pull_request_number}}'
      build.appstudio.redhat.com/target_branch: '{{target_branch}}'
      pipelinesascode.tekton.dev/max-keep-runs: "3"
      pipelinesascode.tekton.dev/on-cel-expression: event == "pull_request" && target_branch == "main" && ("Makefile".pathChanged() || "Dockerfile".pathChanged() || "config/***".pathChanged() || "internal/***".pathChanged() || ".tekton/controller-rhel9-operator-on-pull-request-1-5.yaml".pathChanged())
    creationTimestamp: null
    labels:
      appstudio.openshift.io/application: operator-1-5
      appstudio.openshift.io/component: controller-rhel9-operator-1-5
      pipelines.appstudio.openshift.io/type: build
    name: controller-rhel9-operator-on-pull-request-1-5
    namespace: orchestrator-releng-tenant
  .....
  ```
* Create a PR from the branch created in the first step of this section and
  merge it into the release branch (e.g. `release-1-5`)

#### Update Main branch Via PR

* Navigate to the [.tekton folder](https://github.com/rhdhorchestrator/orchestrator-go-operator/tree/main/.tekton)
  and update the pipeline file names that are suffixed with the incremental (next) release `xxx-1.x.yaml`
  Example assuming 1.6 is the next release.

```console
- controller-rhel9-operator-on-pull-request-1-6.yaml
- controller-rhel9-operator-on-push-1-6.yaml
- orchestrator-operator-bundle-on-pull-request-1-6.yaml
- orchestrator-operator-bundle-on-push-1-6.yaml
```

* In each of the pipeline config file listed above, in the `pipelinesascode.tekton.dev/on-cel-expression`,
  ensure the `target_branch` points to main.
* Update the labels, name, component and any relevant change in each pipeline config.

These changes should be done via PR branch and merged into `main` branch
