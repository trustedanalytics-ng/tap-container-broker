# TAP Container Broker
Container Broker is a microservice developed to be a part of TAP platform.
It monitors the rabbitMQ queue and executes requests for spawning new container instances (such as application or service instances).

## REQUIREMENTS

### Binary
Container Broker depends on Template Repository, Catalog and rabbitMQ components.

### Compilation
* git (for pulling repository)
* go >= 1.6

## Compilation
To build project:
```
  git clone https://github.com/intel-data/tap-container-broker
  cd tap-container-broker
  make build_anywhere
```
Binaries are available in ./application directory.

## USAGE

To provide IP and port for the application, you have to setup environment variables:
```
export BIND_ADDRESS=127.0.0.1
export PORT=80
```

As Container Broker depends on external components, following environment variables should be set:
* CATALOG_KUBERNETES_SERVICE_NAME
* CATALOG_USER
* CATALOG_PASS
* TEMPLATE_REPOSITORY_KUBERNETES_SERVICE_NAME
* TEMPLATE_REPOSITORY_USER
* TEMPLATE_REPOSITORY_PASS
* QUEUE_KUBERNETES_SERVICE_NAME
* QUEUE_USER
* QUEUE_PASS

Other required environment variables:
* MEMORY_LIMIT
MEMORY_LIMIT defines default memory limit for client application. If it's not set, value "0" will be assumed.

Optional:
These variables define default java parameters placed in JAVA_OPTS environment variable for client java applications.

* JAVA_XMS
* JAVA_XMX
* JAVA_METASPACE

Volume optional parameters (currently only for ceph volume types):

* CEPH_FS_TYPE
* CEPH_POOL
* CEPH_MONITORS
* CEPH_SECRET_NAME
* CEPH_USER
* CEPH_IMAGE_SIZE_MB

Monitoring of kubernetes state optional parameters:

* WAIT_FOR_SCALED_REPLICAS_INTERVAL

Container-broker other optional parameters:

* LOGS_TAIL_LINES - maximum lines of logs returned by "/logs" endpoint (500 by default)

Container Broker endpoints are documented in swagger.yaml file.
Below you can find sample Container Broker usage.

#### Create/Delete/Scale instance
To trigger one of the above action proper message [models/serviceInstanceRequest.go](models/serviceInstanceRequest.go) 
has to be sent to one specific the queue and routing key: [models/queue.go](models/queue.go) 

#### Get instance envs
```
curl http://127.0.0.1/api/v1/service/174db3bc-d12a-4fad-63f5-56df07c41325/envs --user admin:password -v
[{"RcName":"x174db3bcd12a4","Containers":[{"Name":"k-logstash14","Envs":{"MANAGED_BY":"TAP","RABBITMQ_PASSWORD":"$random2","RABBITMQ_USERNAME":"$random1","TAP_K8S":"true"}}]}]
```

#### Get secret
```
curl http://127.0.0.1/api/v1/secret/minio --user admin:password
{"metadata":{"name":"minio","namespace":"default","selfLink":"/api/v1/namespaces/default/secrets/minio","uid":"300c0cf1-6ea9-11e6-922b-000c293aca75","resourceVersion":"360","creationTimestamp":"2016-08-30T11:59:21Z"},"data":{"access-key":"c2VjcmV0X2FjY2Vzc19rZXk=","secret-key":"c3VwZXItc2VjcmV0LU1JTkkwLXNlcnZlci1rZXk="},"type":"Opaque"}
```

#### Expose instance
Expose is available for both services types:
* native service instances - only 'ports' and 'hostname' params have to be pass. Ingress object will be created for corresponding kubernetes offering Service
* service instance from service-broker - all params are required ('ports', 'ip', 'hostname'). Ingress, Service and Endpoint objects 
will be created in Kubernetes for corresponding Ip adress

Exposed address pattern: $instanceName-$port.$hostname.$domain

Exposed addresses will ba added to Instance metadata on key 'urls'.
```bash
curl http://127.0.0.1/api/v1/service/:instanceId/expose -X POST -d '{"ports": [2480, 2424], "hostname": "daily", "ip": "10.0.0.43"}'
```

response:
```json
[
    "orient-2480.daily.gotapaas.eu",
    "orient-2424.daily.gotapaas.eu"
]
```

#### Hide instance
```bash
curl http://127.0.0.1/api/v1/service/:instanceId/expose -XDELETE'
```

#### Bind srcInstance to other dstInstance

#### environments bind
Bind operation put into dstInstance containers environments from srcInstance usage srcInstance name as prefix e.g.:

srcInstance:
* name: blue:
* envs: INTEL=inside

in dstInstance we get following env: BLUE_INTEL=inside

#### virtual disk bind
If srcInstance has following prefixed environments:
* EXP_MOUNT_CONFIG_MAP_NAME_ - e.g. EXP_MOUNT_CONFIG_MAP_NAME_MYSQL=mysql-config-map
* EXP_MOUNT_PATH_            - e.g. EXP_MOUNT_PATH_MYSQL=/tmp/msql
* EXP_MOUNT_READ_ONLY_       - e.g. EXP_MOUNT_READ_ONLY_MYSQL=false

Then in dstInstance 'mysql-config-map' will be mount on '/tmp/msql' path with Read/Write access

```bash
curl http://127.0.0.1/api/v1/bind/:srcInstanceId/:dstInstanceId -XPOST'
curl http://127.0.0.1/api/v1/unbind/:srcInstanceId/:dstInstanceId -XPOST'
```

# Persistent volumes

In order to create persistent volume for processed offering, container-broker
assumes that in template for deployment specific annotations are set:

```
"template": {
              "metadata": {                
                "labels": {
                  "idx_and_short_instance_id": "$idx_and_short_instance_id",                  
                },
                "annotations":{
                  "volume_read_only": "false",
                  "volume_size_mb": "102400",
                  "volume_name": "mysql56-persistent-storage"
                }
              }
              }
 ```
 
 Currently only one persistent volume allocation per deployment can be made
 on demand. If your offering requires more volumes then another deployment (pod)
 definition should be created with proper annotation.
 By default volume is read only, has 200mb storage alocated. Name should be same as defined in volue mounts.