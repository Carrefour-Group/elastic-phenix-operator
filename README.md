<p align="center"><img src="logo.png" width="200x" /></p>
<p><div align="center"><h1>Elasticsearch Phenix Operator</h1></div></p>
<p align="center">
<a href="https://goreportcard.com/report/github.com/Carrefour-Group/elastic-phenix-operator"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/Carrefour-Group/elastic-phenix-operator" /></a>
<a href="https://github.com/Carrefour-Group/elastic-phenix-operator/releases/tag/v1.2.0"><img alt="Release" src="https://img.shields.io/badge/release-v1.2.0-blue" /></a>
<a href="https://hub.docker.com/r/carrefourphx/elastic-phenix-operator/tags?page=1&ordering=last_updated"><img alt="Docker" src="https://img.shields.io/badge/docker-tags-orange" /></a>
<a href="https://opensource.org/licenses/Apache-2.0"><img alt="License" src="https://img.shields.io/badge/License-Apache%202.0-green.svg" /></a>
</p>

# Overview

`Elasticsearch Phenix Operator` is a kubernetes operator to manage `elasticsearch` Indices and Templates lifecycle.

Supported Elasticsearch versions are:
  *  Elasticsearch 8+
  *  Elasticsearch 7+
  *  Elasticsearch 6+

See the [Quickstart](https://github.com/Carrefour-Group/elastic-phenix-operator#quick-start) to get started with `Elasticsearch Phenix Operator`.

# Contents
- [Features](#features)
- [Kubernetes Domain, Group and Kinds](#kubernetes-domain-group-and-kinds)
- [Quick Start](#quick-start)
  * [Creating a kubernetes cluster](#creating-a-kubernetes-cluster)
  * [Install prerequisites](#install-prerequisites)
  * [Creating an elasticsearch cluster](#creating-an-elasticsearch-cluster)
  * [Install Elasticsearch Phenix Operator](#install-elasticsearch-phenix-operator)
  * [Creating a secret for connection URL](#creating-a-secret-for-connection-url)
  * [Creating an elasticindex](#creating-an-elasticindex)
  * [Creating an elastictemplate](#creating-an-elastictemplate)
  * [Creating an elasticpipeline](#creating-an-elasticpipeline)
  * [Get created objects and debugging](#get-created-objects-and-debugging)
  * [Deleting elasticindex, elastictemplate with annotation](#deleting-elasticindex-elastictemplate-elasticpipeline-with-annotation)
- [Architecture](#architecture)
- [Operator arguments](#operator-arguments)
- [Release artifacts](#release-artifacts)
- [Validations](#validations)
  * [Syntactic validation](#syntactic-validation)
  * [Semantic validation](#semantic-validation)
    + [on creation](#on-creation)
    + [on update](#on-update)
    + [on delete](#on-delete)
- [Mutation](#mutation)
- [Add new kind to Elasticsearch Phenix Operator](#add-new-kind-to-elasticsearch-phenix-operator)

# Features:
    
  *  Manage Elasticsearch indices, ingest pipelines and templates lifecycle: create, update and delete
  *  Create new indices/templates/pipelines, or manage existing indices/templates/pipelines. In case of existing indices/templates/pipelines, the `ElasticIndex`/`ElasticTemplate`/`ElasticPipeline` object definition should be compatible with existing `index`/`template`/`pipeline`, otherwise you will get a kubernetes object created with `Error` status
  *  One instance of the operator can manage indices, ingest pipelines and templates on different elasticsearch servers
  *  Elasticsearch server URI is provided from a secret when you create ElasticIndex and ElasticTemplate and ElasticPipeline objects
  *  Manage indices and templates and pipelines uniqueness inside kubernetes
  *  A ValidatingWebhook is implemented to validate ElasticIndex and ElasticTemplate and ElasticPipeline objects

# Kubernetes Domain, Group and Kinds

**Domain:** `carrefour.com`

**Group:** `elastic`

**Kinds:** three kinds are available

- `ElasticIndex`: manage elasticsearch indices lifecycle `create`, `update` and `delete`
- `ElasticTemplate`: manage elasticsearch templates lifecycle `create`, `update` and `delete`
- `ElasticPipeline`: manage elasticsearch ingest pipelines lifecycle `create`, `update` and `delete`

# Quick Start

## Creating a kubernetes cluster

You can use `kind` to run a kubernetes cluster in your machine. For more information: https://kind.sigs.k8s.io/docs/user/quick-start/

Create a cluster:

```
kind create cluster --image=kindest/node:v1.17.0
```

## Install prerequisites

`Cert-manager` is needed to handle TLS certificate for admission webhook servers. You need `cert-manager` version `v1.0.0` or above. For more information: https://github.com/jetstack/cert-manager/

To install `cert-manager`:

```
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.1/cert-manager.yaml
```

You should wait until the `cert-manager` becomes in running state:

```
kubectl wait --for=condition=Ready --timeout=-1s --all pods -n cert-manager
```

You can use `ECK` (Elastic Cloud on Kubernetes) to create an elasticsearch cluster in kubernetes. For more information: https://www.elastic.co/guide/en/cloud-on-k8s/master/k8s-overview.html

To install `ECK`:

```
kubectl apply -f https://download.elastic.co/downloads/eck/1.3.0/all-in-one.yaml
```

You should wait until the `ECK` operator becomes in running state:

```
kubectl wait --for=condition=Ready --timeout=-1s --all pods -n elastic-system
```

## Creating an elasticsearch cluster

You can create a single node `Elasticsearch` cluster:

```
cat <<EOF | kubectl apply -n elastic-system -f -
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: elastic-system
spec:
  version: 7.9.2
  nodeSets:
  - name: default
    count: 1
    config:
      node.master: true
      node.data: true
      node.ingest: true
      node.store.allow_mmap: false
      xpack.security.authc:
        anonymous:
          username: anonymous_user
          roles: superuser
          authz_exception: false
  http:
    service:
      spec:
        type: ClusterIP
    tls:
      selfSignedCertificate:
        disabled: true
        subjectAltNames:
          - dns: localhost,127.0.0.1
EOF
```

You should wait until the `elasticsearch` cluster becomes in running state:

```
kubectl wait --for=condition=Ready --timeout=-1s --all pods -n elastic-system
```

## Install Elasticsearch Phenix Operator

To install `Elasticsearch Phenix Operator` (`EPO`):

```
kubectl apply -f https://raw.githubusercontent.com/Carrefour-Group/elastic-phenix-operator/v1.2.0/manifests/epo-all-in-one.yaml
```

You should wait until `Elasticsearch Phenix Operator` becomes in running state:

```
kubectl wait --for=condition=Ready --timeout=-1s --all pods -n elastic-phenix-operator-system
```

To access logs for deployment:

```
kubectl logs deployment/elastic-phenix-operator-controller-manager -c manager -n elastic-phenix-operator-system
```

## Creating a secret for connection URL

You can find samples located at `config/samples`.

Before creating an `ElasticIndex` or an `ElasticTemplate`, you should create a secret containing elasticsearch uri that respects this pattern: `<scheme>://<user>:<password>@<hostname>:<port>` e.g. `http://localhost:9200`, `https://elastic:pass@myshost:9200`

```
cat <<EOF | kubectl apply -n elastic-phenix-operator-system -f -
apiVersion: v1
kind: Secret
metadata:
  name: elasticsearch-cluster-secret
  namespace: elastic-phenix-operator-system
type: Opaque
stringData:
  uri: http://elastic-system-es-http.elastic-system.svc:9200
EOF
```

## Creating an elasticindex

When creating an `ElasticIndex`, you should reference the elasticsearch server URI from the secret created before:
**/!\\ Secret should be in the same namespace, otherwise you will get an error /!\\**

```
cat <<EOF | kubectl apply -n elastic-phenix-operator-system -f -
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticIndex
metadata:
  name: product-index
  namespace: elastic-phenix-operator-system
spec:
  indexName: product
  elasticURI:
    secretKeyRef:
      name: elasticsearch-cluster-secret
      key: uri
  numberOfShards: 6
  numberOfReplicas: 1
  model: |-
    {
      "settings": {
        "index.codec": "best_compression"
      },
      "mappings": {
        "_source": {
          "enabled": true
        },
        "dynamic": false,
        "properties": {
          "barcode": {
            "type": "keyword",
            "index": true
          },
          "description": {
            "type": "text",
            "index": true
          }
        }
      }
    }
EOF
```

## Creating an elastictemplate

When creating an `ElasticTemplate`, you should reference the elasticsearch server URI from the secret created before:
**/!\\ Secret should be in the same namespace, otherwise you will get an error /!\\**

```
cat <<EOF | kubectl apply -n elastic-phenix-operator-system -f -
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticTemplate
metadata:
  name: invoice-template
  namespace: elastic-phenix-operator-system
spec:
  templateName: invoice
  elasticURI:
    secretKeyRef:
      name: elasticsearch-cluster-secret
      key: uri
  numberOfShards: 5
  numberOfReplicas: 1
  order: 1
  model: |-
    {
      "index_patterns": ["invoice*"],
      "settings": {
        "index.codec": "best_compression"
      },
      "mappings": {
        "_source": {
          "enabled": true
        },
        "properties": {
          "key": {
            "type": "keyword",
            "index": true
          },
          "content": {
            "type": "text",
            "index": true
          }
        }
      }
    }
EOF
```

## Creating an elasticpipeline

When creating an `ElasticPipeline`, you should reference the elasticsearch server URI from the secret created before:
**/!\\ Secret should be in the same namespace, otherwise you will get an error /!\\**

```
cat <<EOF | kubectl apply -n elastic-phenix-operator-system -f -
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticPipeline
metadata:
  name: test-pipeline
  namespace: elastic-phenix-operator-system
spec:
    pipelineName: "test-pipeline"
  elasticURI:
    secretKeyRef:
      name: elasticsearch-cluster-secret
      key: uri
  model: |-
    {
      "description": "My optional pipeline description",
      "processors": [
        {
          "set": {
            "description": "My optional processor description",
            "field": "my-long-field",
            "value": 10
          }
        },
        {
          "set": {
            "description": "Set 'my-boolean-field' to true",
            "field": "my-boolean-field",
            "value": true
          }
        },
        {
          "lowercase": {
            "field": "my-keyword-field"
          }
        }
      ]
    }
EOF
```

## Get created objects and debugging

To get created object, you can use `kubectl` cli:

```
> kubectl get elasticindex -n elastic-phenix-operator-system

NAME            INDEX_NAME   SHARDS   REPLICAS   STATUS    AGE
product-index   product      6        1          Created   24m
city-index      city         4        3          Error     21m


> kubectl get elastictemplate -n elastic-phenix-operator-system

NAME                TEMPLATE_NAME   SHARDS   REPLICAS   STATUS    AGE
invoice-template    invoice         5        3          Created   9m


> kubectl get elasticpipeline -n elastic-phenix-operator-system

NAME                     PIPELINE_NAME   STATUS    AGE
elasticpipeline-sample   test-pipeline   Created   9h

```


You can also check indices and templates in `elasticsearch` cluster:

```
kubectl exec -it pod/elastic-system-es-default-0 -n elastic-system -- curl "localhost:9200/_cat/indices/product?v"
kubectl exec -it pod/elastic-system-es-default-0 -n elastic-system -- curl "localhost:9200/_cat/templates/invoice?v"
kubectl exec -it pod/elastic-system-es-default-0 -n elastic-system -- curl "localhost:9200/_ingest/pipeline/test-pipeline"
```

The `STATUS` column indicates whether `index`/`template`/`pipeline` was created successfully in elasticsearch server. Possible values: 

  * `Created`: when `index`/`template` was created successfully in elasticsearch server
  * `Error`, `Retry`: when error has occurred during creating or updating an `elasticindex`/`elastictemplate`/`elasticpipeline`

When you have an `elasticindex`/`elastictemplate`/`elasticpipeline` with `Error` or `Retry` status, use `kubectl describe` to get more details:

```
> kubectl describe elasticindex/city-index -n elastic-phenix-operator-system

Name:         city-index
Namespace:    elastic-phenix-operator-system
Annotations:  API Version:  elastic.carrefour.com/v1alpha1
Kind:         ElasticIndex
Metadata:
  ...
Spec:
  ...
Status:
  Http Code Status:  400
  Message:           [400 Bad Request] {"error":{"root_cause":[{"type":"mapper_parsing_exception","reason":"Root mapping definition has unsupported parameters:  [dynamicc : false]"}],"type":"mapper_parsing_exception","reason":"Failed to parse mapping: Root mapping definition has unsupported parameters:  [dynamicc : false]","caused_by":{"type":"mapper_parsing_exception","reason":"Root mapping definition has unsupported parameters:  [dynamicc : false]"}},"status":400}
  Status:            Error
```

## Deleting elasticindex, elastictemplate, elasticpipeline with annotation

When you delete an `ElasticIndex`/`ElasticTemplate` kubernetes object, the `index`/`template`/`pipeline` in `elasticsearch` cluster will remain existing.

```
kubectl delete elastictemplate/invoice-template -n elastic-phenix-operator-system
kubectl delete elasticindex/product-index -n elastic-phenix-operator-system
kubectl delete elasticpipeline/test-pipeline -n elastic-phenix-operator-system
```

If you want to delete the `index`/`template`/`pipeline` in `elasticsearch` cluster too, you should add the annotation `carrefour.com/delete-in-cluster=true` to your kubernetes object.

```
kubectl annotate elastictemplate/invoice-template carrefour.com/delete-in-cluster=true -n elastic-phenix-operator-system
kubectl annotate elasticindex/product-index carrefour.com/delete-in-cluster=true -n elastic-phenix-operator-system
kubectl annotate elasticpipeline/test-pipeline/delete-in-cluster=true -n elastic-phenix-operator-system
```

Now, when you delete your `ElasticIndex`/`ElasticTemplate`/`ElasticPipeline` kubernetes object, elasticsearch `index`/`template`/`pipeline` will be deleted too from `elasticsearch` cluster.

**/!\\ For indices deletion, you will lose indices data in elasticsearch cluster /!\\**

**/!\\ For pipelines deletion, the delete will not work if some index in elasticsearch uses the ingest pipeline as a default_pipeline or final_pipeline /!\\**

# Architecture

![elastic-phenix-operator](elastic-phenix-operator.png)

# Operator arguments

You can customise `Elasticsearch Phenix Operator` behavior using these `manager` arguments:

- `namespaces`: create a cache on namespaces and watch only these namespace (defaults to all namespaces)
- `namespaces-regex-filter`: watch all namespaces and filter before reconciliation process (defaults to no filter applied)

# Release artifacts

When releasing `Elasticsearch Phenix Operator`, two artifacts are generated:

- a docker image containing `elastic-phenix-operator` manager embedding `ElasticIndex` and `ElasticTemplate` controllers. All docker images are published in docker hub: https://hub.docker.com/r/carrefourphx/elastic-phenix-operator
- an all-in-one kubernetes manifest file located at `manifest/epo-all-in-one.yaml` that defines all kubernetes objects needed to install and run the `Elasticsearch Phenix Operator`: `CustoResourceDefinition`, `Namespace`, `Deployment`, `Service`, `MutatingWebhookConfiguration`, `ValidatingWebhookConfiguration`, `Role`, `ClusterRole`, `RoleBinding`, `ClusterRoleBinding`, `Certificate`

# Validations

`ElasticIndex` and `ElasticTemplate` kubernetes objects creation goes through two steps of validation: **syntactic validation** and **semantic validation**

## Syntactic validation

A syntactic validation is defined in `CustomResourceDefinition` (section `openAPIV3Schema`). 

These rules are defined:

- `indexName` and `templateName` fields are mandatory, and value should respect regex `^[a-z0-9-_\.]+$`
- `numberOfShards` field is mandatory, and value should be between 1 and 500
- `numberOfReplicas` field is mandatory, and value should be between 1 and 3
- `model` field is mandatory
- `elasticURI` field is mandatory

## Semantic validation

A semantic validation is defined in a kubernetes `ValidatingWebhook`. 

Multiple rules are implemented for different actions: `create`, `update` or `delete`

### on creation

- `model` field content is a valid json
- `ElasticIndex model` json root content contains at most `aliases`, `mappings`, `settings` 
- `ElasticTemplate model` json root content contains at most `aliases`, `mappings`, `settings`, `index_patterns`, `version`
- `ElasticTemplate` model field contains the mandatory field `index_patterns`
- `elasticURI` secret should exist on the same `ElasticIndex`/`ElasticTemplate` namespace
- `elasticURI` secret should respect this pattern: `<scheme>://<user>:<password>@<hostname>:<port>` e.g. `http://localhost:9200`, `https://elastic:pass@myshost:9200`
- manage index and template **uniqueness**: you cannot create the same elasticsearch index/template (`indexName`/`templateName` field) on different kubernetes `ElasticIndex`/`ElasticTemplate` objects when you specify the same elasticsearch `host:port` in `elasticURI` secret


### on update

`ElasticIndex`: you cannot update 
- `indexName` field
- `numberOfShards` field 
- `model` settings (only `numberOfReplicas` update is allowed)

`ElasticTemplate`: you cannot update
- `templateName` field
- `model` field if new model content is not a valid json

`ElasticPipeline`: you cannot update
- `pipelineName` field
- `model` field if new model content is not a valid json

For both `ElasticIndex`/`ElasticTemplate`/`ElasticPipeline` when updating `elasticURI` secret:
- it should exist on the same `ElasticIndex`/`ElasticTemplate`/`ElasticPipeline` namespace
- it should respect this pattern: `<scheme>://<user>:<password>@<hostname>:<port>` e.g. `http://localhost:9200`, `https://elastic:pass@myshost:9200`
- you cannot update elasticsearch `host:port`, only `user` and/or `password` can be updated in `elasticURI` content

### on delete

- `elasticURI` secret should exists on the same `ElasticIndex`/`ElasticTemplate`/`ElasticPipeline` namespace

# Mutation

A `MutatingWebhook` is implemented to initialize `numberOfShards` and `numberOfReplicas` settings fields, from fields `numberOfShards` and `numberOfReplicas` of an `ElasticIndex`/`ElasticTemplate`.

If user has defined `numberOfShards` or/and `numberOfReplicas` in settings in `model` field, **these values will be overridden** by `numberOfShards` and `numberOfReplicas` fields in the `ElasticIndex`/`ElasticTemplate` defintion.

For this `ElasticIndex` defintion:

```
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticIndex
metadata:
  name: product-index
spec:
  indexName: product
  elasticURI:
    secretKeyRef:
      name: elasticsearch-cluster-secret
      key: uri
  numberOfShards: 6
  numberOfReplicas: 1
  model: |-
    {
      "settings": {
        "numberOfReplicas": 3
      },
      "mappings": {
        "_source": {
          "enabled": true
        },
        "dynamic": false,
        "properties": {
          "barcode": {
            "type": "keyword",
            "index": true
          },
          "description": {
            "type": "text",
            "index": true
          }
        }
      }
    }
```

=> The result of mutation step will be:

```
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticIndex
metadata:
  name: product-index
spec:
  indexName: product
  elasticURI:
    secretKeyRef:
      name: elasticsearch-cluster-secret
      key: uri
  numberOfShards: 6
  numberOfReplicas: 1
  model: |-
    {
      "settings": {
        "numberOfReplicas": 1
        "numberOfShards": 6
      },
      "mappings": {
        "_source": {
          "enabled": true
        },
        "dynamic": false,
        "properties": {
          "barcode": {
            "type": "keyword",
            "index": true
          },
          "description": {
            "type": "text",
            "index": true
          }
        }
      }
    }
```

# Add new kind to Elasticsearch Phenix Operator

This operator was generated using `kubebuilder 2.3.1`. For more details about `kubebuiler`: https://book.kubebuilder.io/

Let's say that you want to add a new `Kind` to manage elasticsearch pipelines: `ElasticPipeline`

You should run these commands:

```
kubebuilder create api --group elastic --version v1alpha1 --kind ElasticPipeline

kubebuilder create webhook --group elastic --version v1alpha1 --kind ElasticPipeline --defaulting
```
