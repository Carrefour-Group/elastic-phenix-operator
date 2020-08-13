# elastic-phenix-operator

`elastic-phenix-operator` is a kubernetes operator to manage elasticsearch Indices and Templates.

# Kubernetes Domain, Group and Kinds

**Domain:** `carrefour.com`

**Group:** `elastic`

**Kinds:** two kinds are available

- `ElasticIndex`: manage elasticsearch indices lifecycle `create`, `update` and `delete`
- `ElasticTemplate`: manage elasticsearch templates lifecycle `create`, `update` and `delete`

# Architecture

![elastic-phenix-operator](elastic-phenix-operator.png)

# Operator features

`elastic-phenix-operator` is compatible with `elasticsearch 6` and `elasticsearch 7` servers.

It is possible to create `ElasticIndex` and `ElasticTemplate` objects:

- for new `index` / `template`
- on existing `index` / `template`, in this case your kubernetes object defintion should be compatible with existing `index` / `template`, otherwise you will get a kuberenetes object created with `Error` status

To protect sensitive data, elasticsearch server URI should be provided from a secret when creating `ElasticIndex` and `ElasticTemplate` objects.

# Operator arguments

You can customise `elastic-phenix-operator` behavior using these `manager` arguments:

- `enable-delete`: enable `indices`/`templates` deletion on elasticsearch server when kubernetes `elasticindex`/`elastictemplate` objects are deleted. If not specified, only kubernetes `elasticindex`/`elastictemplate` objects will be deleted
- `namespaces`: create a cache on namespaces and watch only these namespace (defaults to all namespaces)
- `namespaces-regex-filter`: watch all namespaces and filter before reconciliation process (defaults to no filter applied)

# Release artifacts

When releasing `elastic-phenix-operator`, two artifacts are released:

- a docker image containing `elastic-phenix-operator` manager embedding `ElasticIndex` and `ElasticTemplate` controllers.
- an all-in-one kubernetes manifest file located at `manifest/epo-all-in-one.yaml` that defines all kubernetes objects needed to install and run the `elastic-phenix-operator`: `CustoResourceDefinition`, `Namespace`, `Deployment`, `Service`, `MutatingWebhookConfiguration`, `ValidatingWebhookConfiguration`, `Role`, `ClusterRole`, `RoleBinding`, `ClusterRoleBinding`, `Certificate`

# Create your first ElasticIndex and ElasticTemplate

You can find samples located at `config/samples`.

Before creating an `ElasticIndex` or an `ElasticTemplate`, you should create a secret containing elasticsearch uri that respects this pattern: `<scheme>://<user>:<password>@<hostname>:<port>` e.g. `http://localhost:9200`, `https://elastic:pass@myshost:9200`

```
apiVersion: v1
kind: Secret
metadata:
  name: elasticsearch-cluster-secret
  namespace: elasticsearch-dev
type: Opaque
stringData:
  uri: http://elastic:mypass@elasticsearchhost:9200
```

Then create an `ElasticIndex` and you should reference the elasticsearch server URI from the secret created before:
**/!\\ Secret should be in the same namespace, otherwise you will get an error /!\\**

```
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticIndex
metadata:
  name: product-index
  namespace: elasticsearch-dev
spec:
  indexName: product
  elasticURI:
    secretKeyRef:
      key: uri
      name: elasticsearch-cluster-secret
  numberOfShards: 6
  numberOfReplicas: 1
  model: |-
    {
      "settings": {
      },
      "mappings": {
        "_doc": {
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
    }
```

An example of an `ElasticTemplate`:

```
apiVersion: elastic.carrefour.com/v1alpha1
kind: ElasticTemplate
metadata:
  name: invoice-template
  namespace: elasticsearch-dev
spec:
  templateName: invoice
  elasticURI:
    secretKeyRef:
      key: uri
      name: elasticsearch-cluster-secret
  numberOfShards: 5
  numberOfReplicas: 3
  model: |-
    {
      "index_patterns": ["invoice*"],
      "settings": {
      },
      "mappings": {
        "_doc": {
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
    }
```

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

For both `ElasticIndex`/`ElasticTemplate` when updating `elasticURI` secret:
- it should exist on the same `ElasticIndex`/`ElasticTemplate` namespace
- it should respect this pattern: `<scheme>://<user>:<password>@<hostname>:<port>` e.g. `http://localhost:9200`, `https://elastic:pass@myshost:9200`
- you cannot update elasticsearch `host:port`, only `user` and/or `password` can be updated in `elasticURI` content

### on delete

- `elasticURI` secret should exists on the same `ElasticIndex`/`ElasticTemplate` namespace

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
        "_doc": {
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
        "_doc": {
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
    }
```

# Get created objects and debugging

To get created object, you can use `kubectl` cli:

```
> kubectl get elasticindex -n mynamespace

NAME            INDEX_NAME   SHARDS   REPLICAS   STATUS    AGE
product-index   product      6        1          Created   24m
city-index      city         4        3          Error     21m


> kubectl get elastictemplate -n mynamespace

NAME                TEMPLATE_NAME   SHARDS   REPLICAS   STATUS    AGE
invoice-template    invoice         5        3          Created   9m
```

When you have an `elasticindex`/`elastictemplate` with `Error` status, use `kubectl describe` to get more details:

```
> kubectl describe elasticindex/city-index -n mynamespace

Name:         city-index
Namespace:    elasticsearch-dev
Annotations:  API Version:  elastic.carrefour.com/v1alpha1
Kind:         ElasticIndex
Metadata:
  ...
Spec:
  ...
Status:
  Http Code Status:  400
  Message:           [400 Bad Request] {"error":{"root_cause":[{"type":"mapper_parsing_exception","reason":"Root mapping definition has unsupported parameters:  [dynamicc : false]"}],"type":"mapper_parsing_exception","reason":"Failed to parse mapping [_doc]: Root mapping definition has unsupported parameters:  [dynamicc : false]","caused_by":{"type":"mapper_parsing_exception","reason":"Root mapping definition has unsupported parameters:  [dynamicc : false]"}},"status":400}
  Status:            Error
```

# Add new kind to elastic-phenix-operator

This operator was generated using `kubebuilder 2.3.1`. For more details about `kubebuiler`: https://book.kubebuilder.io/

Let's say that you want to add a new `Kind` to manage elasticsearch pipelines: `ElasticPipeline`

You should run these commands:

```
kubebuilder create api --group elastic --version v1alpha1 --kind ElasticPipeline

kubebuilder create webhook --group elastic --version v1alpha1 --kind ElasticPipeline --defaulting
```
