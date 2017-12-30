# Fargate CLI

![](doc/media/fargate.png)

Deploy serverless containers onto the cloud from your command line

## Usage

### Configuration

#### Region

By default, fargate uses *us-east-1* as this is the single region where AWS
Fargate is available. The CLI accepts a --region parameter for future use and
will honor *AWS_REGION* and *AWS_DEFAULT_REGION* environment settings. Note that
specifying a region where all required services aren't available will return an
error.

See the [Region Table][region-table] for a breakdown of what services are
available in which regions.

#### Credentials

fargate is built using the [AWS SDK for Go][go-sdk] which looks for credentials
in the following locations:

1. [Environment Variables][go-env-vars]

1. [Shared Credentials File][go-shared-credentials-file]

1. [EC2 Instance Profile][go-iam-roles-for-ec2-instances]

For more information see [Specifying Credentials][go-specifying-credentials] in
the AWS SDK for Go documentation.

### Commands

- [Tasks](#tasks)
- [Services](#services)
- [Load Balancers](#load-balancers)
- [Certificates](#certificates)

#### Global Flags

| Flag | Default | Description |
| --- | --- | --- |
| --cluster | fargate | ECS cluster name (default "fargate") |
| --region | us-east-1 | AWS region (default "us-east-1") |
| --no-color | false | Disable color output |
| --verbose | false | Verbose output |

#### Tasks

Tasks are one-time executions of your container. Instances of your task are run
until you manually stop them either through AWS APIs, the AWS Management
Console, or `fargate task stop`, or until they are interrupted for any reason.

- [list](#fargate-task-list)
- [run](#fargate-task-run)
- [info](#fargate-task-info)
- [ps](#fargate-task-ps)
- [logs](#fargate-task-logs)
- [stop](#fargate-task-stop)

##### fargate task list

```console
fargate task list
```

List running task groups

##### fargate task run

```console
fargate task run <task-group-name> [--num <count>] [--cpu <cpu-units>] [--memory <MiB>]
                                   [--image <docker-image>] [--env <key=value>]
```

Run new tasks

You must specify a task group name in order to interact with the task(s) in
subsequent commands to view logs, stop and inspect tasks. Task group names do
not have to be unique -- multiple configurations of task instances can be
started with the same task group.

Multiple instances of a task can be run by specifying a number in the --num
flag. If no number is specified, a single task instance will be run.

CPU and memory settings can be optionally specified as CPU units and mebibytes
respectively using the --cpu and --memory flags. Every 1024 CPU units is
equivilent to a single vCPU. AWS Fargate only supports certain combinations of
CPU and memory configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

If not specified, fargate will launch minimally sized tasks at 0.25 vCPU (256
CPU units) and 0.5GB (512 MiB) of memory.

The Docker container image to use in the task can be optionally specified via
the --image flag. If not specified, fargate will build a new Docker container
image from the current working directory and push it to Amazon ECR in a
repository named for the task group. If the current working directory is a git
repository, the container image will be tagged with the short ref of the HEAD
commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.

##### fargate task info

```console
fargate task info <task-group-name> [--task <task-id>]
```

Inspect tasks

Shows extended information for each running task within a task group or for
specific tasks specified with the --task flag. Information includes environment
variables which could differ between tasks in a task group. To inspect multiple
specific tasks within a task group specific --task with a task ID multiple
times.

##### fargate task ps

```console
fargate task ps <task-group-name>
```

List running tasks

##### fargate task logs

```console
fargate task logs <task-group-name> [--follow] [--start <time-expression>] [--end <time-expression>]
                                    [--filter <filter-expression>] [--task <task-id>]
```

Show logs from tasks

Return either a specific segment of task logs or tail logs in real-time using
the --follow option. Logs are prefixed by their log stream name which is in the
format of "fargate/\<task-group-name>/\<task-id>."

Follow will continue to run and return logs until interrupted by Control-C. If
--follow is passed --end cannot be specified.

Logs can be returned for specific tasks within a task group by passing a task
ID via the --task flag. Pass --task with a task ID multiple times in order to
retrieve logs from multiple specific tasks.

A specific window of logs can be requested by passing --start and --end options
with a time expression. The time expression can be either a duration or a
timestamp:

  - Duration (e.g. -1h [one hour ago], -1h10m30s [one hour, ten minutes, and
    thirty seconds ago], 2h [two hours from now])
  - Timestamp with optional timezone in the format of YYYY-MM-DD HH:MM:SS [TZ];
    timezone will default to UTC if omitted (e.g. 2017-12-22 15:10:03 EST)

You can filter logs for specific term by passing a filter expression via the
--filter flag. Pass a single term to search for that term, pass multiple terms
to search for log messages that include all terms. See the [CloudWatch Logs
documentation][cwl-filter-expression] for more details.

##### fargate task stop

```console
fargate task stop <task-group-name> [--task <task-id>]
```

Stop tasks

Stops all tasks within a task group if run with only a task group name or stops
individual tasks if one or more tasks are passed via the --task flag. Specify
--task with a task ID parameter multiple times to stop multiple specific tasks.

#### Services

Services manage long-lived instances of your containers that are run on AWS
Fargate. If your container exits for any reason, the service scheduler will
restart your containers and ensure your service has the desired number of
tasks running. Services can be used in concert with a load balancer to
distribute traffic amongst the tasks in your service.

- [list](#fargate-service-list)
- [create](#fargate-service-create)
- [deploy](#fargate-service-deploy)
- [info](#fargate-service-info)
- [logs](#fargate-service-logs)
- [ps](#fargate-service-ps)
- [scale](#fargate-service-scale)
- [env set](#fargate-service-env-set)
- [env unset](#fargate-service-env-unset)
- [env list](#fargate-service-env-list)
- [update](#fargate-service-update)
- [restart](#fargate-service-restart)
- [destroy](#fargate-service-destroy)

##### fargate service list

```console
fargate service list
```

List services

##### fargate service create

```console
fargate service create <service name> [--cpu <cpu units>] [--memory <MiB>] [--port <port-expression>]
                                      [--lb <load-balancer-name>] [--rule <rule-expression>]
                                      [--image <docker-image>] [--env <key=value>] [--num <count>]
```

Create a new service

CPU and memory settings can be optionally specified as CPU units and mebibytes
respectively using the --cpu and --memory flags. Every 1024 CPU units is
equivilent to a single vCPU. AWS Fargate only supports certain combinations of
CPU and memory configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

If not specified, fargate will launch minimally sized tasks at 0.25 vCPU (256
CPU units) and 0.5GB (512 MiB) of memory.

The Docker container image to use in the service can be optionally specified
via the --image flag. If not specified, fargate will build a new Docker
container image from the current working directory and push it to Amazon ECR in
a repository named for the task group. If the current working directory is a
git repository, the container image will be tagged with the short ref of the
HEAD commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

Services can optionally be configured to use a load balancer. To put a load
balancer in front a service, pass the --lb flag with the name of a load
balancer. If you specify a load balancer, you must also specify a port via the
--port flag to which the load balancer should forward requests. Optionally,
Application Load Balancers can be configured to route HTTP/HTTPS traffic to the
service based upon a rule. Rules are configured by passing one or more rules by
specifying the --rule flag along with a rule expression. Rule expressions are
in the format of TYPE=VALUE. Type can either be PATH or HOST. PATH matches the
PATH of the request and HOST matches the requested hostname in the HTTP
request. Both PATH and HOST types can include up to three wildcard characters:
\* to match multiple characters and ? to match a single character.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.

Specify the desired count of tasks the service should maintain by passing the
--num flag with a number. If you omit this flag, fargate will configure a
service with a desired number of tasks of 1.

##### fargate service deploy

```console
fargate service deploy <service-name> [--image <docker-image>]
```

Deploy new image to service

The Docker container image to use in the service can be optionally specified
via the --image flag. If not specified, fargate will build a new Docker
container image from the current working directory and push it to Amazon ECR in
a repository named for the task group. If the current working directory is a
git repository, the container image will be tagged with the short ref of the
HEAD commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

##### fargate service info

```console
fargate service info <service-name>
```

Inspect service

Show extended information for a service including load balancer configuration,
active deployments, and environment variables.

Deployments show active versions of your service that are running. Multiple
deployments are shown if a service is transitioning due to a deployment or
update to configuration such a CPU, memory, or environment variables.

##### fargate service logs

```console
fargate service logs <service-name> [--follow] [--start <time-expression>] [--end <time-expression>]
                                    [--filter <filter-expression>] [--task <task-id>]
```

Show logs from tasks in a service

Return either a specific segment of service logs or tail logs in real-time
using the --follow option. Logs are prefixed by their log stream name which is
in the format of "fargate/\<service-name>/\<task-id>."

Follow will continue to run and return logs until interrupted by Control-C. If
--follow is passed --end cannot be specified.

Logs can be returned for specific tasks within a service by passing a task ID
via the --task flag. Pass --task with a task ID multiple times in order to
retrieve logs from multiple specific tasks.

A specific window of logs can be requested by passing --start and --end options
with a time expression. The time expression can be either a duration or a
timestamp:

  - Duration (e.g. -1h [one hour ago], -1h10m30s [one hour, ten minutes, and
    thirty seconds ago], 2h [two hours from now])
  - Timestamp with optional timezone in the format of YYYY-MM-DD HH:MM:SS [TZ];
    timezone will default to UTC if omitted (e.g. 2017-12-22 15:10:03 EST)

You can filter logs for specific term by passing a filter expression via the
--filter flag. Pass a single term to search for that term, pass multiple terms
to search for log messages that include all terms. See the [CloudWatch Logs
documentation][cwl-filter-expression] for more details.

##### fargate service ps

```console
fargate service ps <service-name>
```

List running tasks for a service

##### fargate service scale

```console
fargate service scale <service-name> <scale-expression>
```

Scale number of tasks in a service

Changes the number of desired tasks to be run in a service by the given scale
expression. A scale expression can either be an absolute number or a delta
specified with a sign such as +5 or -2.

##### fargate service env set

```console
fargate service env set <service-name> --env <key=value>
```

Set environment variables

At least one environment variable must be specified via the --env flag. Specify
--env with a key=value parameter multiple times to add multiple variables.

##### fargate service env unset

```console
fargate service env unset <service-name> --key <key-name>
```

Unset environment variables

Unsets the environment variable specified via the --key flag. Specify --key with
a key name multiple times to unset multiple variables.

##### fargate service env list

```console
fargate service env list <service-name>
```

Show environment variables

##### fargate service update

```console
fargate service update <service-name> [--cpu <cpu-units>] [--memory <MiB>]
```

Update service configuration

CPU and memory settings are specified as CPU units and mebibytes respectively
using the --cpu and --memory flags. Every 1024 CPU units is equivilent to a
single vCPU. AWS Fargate only supports certain combinations of CPU and memory
configurations:

| CPU (CPU Units) | Memory (MiB)                          |
| --------------- | ------------------------------------- |
| 256             | 512, 1024, or 2048                    |
| 512             | 1024 through 4096 in 1GiB increments  |
| 1024            | 2048 through 8192 in 1GiB increments  |
| 2048            | 4096 through 16384 in 1GiB increments |
| 4096            | 8192 through 30720 in 1GiB increments |

At least one of --cpu or --memory must be specified.

##### fargate service restart

```console
fargate service restart <service-name>
```

Restart service

Creates a new set of tasks for the service and stops the previous tasks. This
is useful if your service needs to reload data cached from an external source,
for example.

##### fargate service destroy

```console
fargate service destroy <service-name>
```

Destroy service

In order to destroy a service, it must first be scaled to 0 running tasks.

#### Load Balancers

Load balancers distribute incoming traffic between the tasks within a service
for HTTP/HTTPS and TCP applications. HTTP/HTTPS load balancers can route to
multiple services based upon rules you specify when you create a new service.

- [list](#fargate-lb-list)
- [create](#fargate-lb-create)
- [destroy](#fargate-lb-destroy)
- [alias](#fargate-lb-alias)
- [info](#fargate-lb-info)

##### fargate lb list

```console
fargate lb list
```

List load balancers

##### fargate lb create

```console
fargate lb create <load-balancer-name> --port <port-expression> [--certificate <certificate-name>]
                                       [--security-group-id <security-group-id>]
```

Create a load balancer

At least one port must be specified for the load balancer listener via the
--port flag and a port expression of protocol:port-number. For example, if you
wanted an HTTP load balancer to listen on port 80, you would specify HTTP:80.
Valid protocols are HTTP, HTTPS, and TCP. You can specify multiple listeners by
passing the --port flag with a port expression multiple times. You cannot mix
TCP ports with HTTP/HTTPS ports on a single load balancer.

You can optionally include certificates to secure HTTPS ports by passed the
--certificate flag along with a certificate name. This option can be specified
multiple times to add additional certificates to a single load balancer which
uses Service Name Identification (SNI) to select the appropriate certificate
for the request.

Security groups can optionally be specified for HTTP/HTTPS load balancers by
passing the --security-group-id flag with a security group ID. To add multiple
security groups, pass --security-group-id with a security group ID multiple
times. If --security-group-id is omitted, a permissive security group will be
applied to the load balancer.

##### fargate lb destroy

```console
fargate lb destroy <load-balancer-name>
```

Destroy load balancer

##### fargate lb alias

```console
fargate lb alias <load-balancer-name> <hostname>
```

Create a load balancer alias record

Create an alias record to the load balancer for domains that are hosted within
Amazon Route 53 and within the same AWS account. If you're using another DNS
provider or host your domains in a differnt account, you will need to manually
create this record.

##### fargate lb info

```console
fargate lb info <load-balancer-name>
```

Inspect load balancer

Returns extended information about a load balancer including a list of
listeners, rules, and certificates in use by the load balancer.


#### Certificates

Certificates are TLS certificates issued by or imported into AWS Certificate
Manager for use in securing traffic between load balancers and end users. ACM
provides TLS certificates free of charge for use within AWS resources.

- [list](#fargate-certificate-list)
- [import](#fargate-certificate-import)
- [request](#fargate-certificate-request)
- [info](#fargate-certificate-info)
- [validate](#fargate-certificate-validate)
- [destroy](#fargate-certificate-destroy)

##### fargate certificate list

```console
fargate certificate list
```

List certificates

##### fargate certificate import

```console
fargate certificate import <domain-name> --certificate <filename> --key <filename> [--chain <filename>]
```

Import a certificate

Upload a certificate from a certificate file, a private key file, an optionally
an intermediate certificate chain file. The files must be PEM-encoded and the
private key must not be encrypted or protected by a passphrase. See the
[AWS Certificate Manager documentation][acm-import-cert] for more details.

##### fargate certificate request

```console
fargate certificate request <domain-name> [--alias <domain-name>]
```

Request a certificate

Certificates can be for a fully qualified domain name (e.g. www.example.com) or
a wildcard domain name (e.g. *.example.com). You can add aliases to a
certificate by specifying additional domain names via the --alias flag. To add
multiple aliases, pass --alias multiple times. By default, AWS Certificate
Manager has a limit of 10 domain names per certificate, but this limit can be
raised by AWS support.

##### fargate certificate info

```console
fargate certificate info <domain-name>
```

Inspect certificate

Show extended information for a certificate including each validation for the
certificate including any DNS records which must be created to validate
domain ownership.

##### fargate certificate validate

```console
fargate certificate validate <domain-name>
```

Validate certificate ownership

fargate will automatically create DNS validation record to verify ownership for
any domain names that are hosted within Amazon Route 53. If your certificate
has aliases, a validation record will be attempted per alias. Any records whose
domains are hosted in other DNS hosting providers or in other DNS accounts
and cannot be automatically validated will have the necessary records output.
These records are also available in `fargate certificate info \<domain-name>`.

AWS Certificate Manager may take up to several hours after the DNS records are
created to complete validation and issue the certificate.

##### fargate certificate destroy

```console
fargate certificate destroy <domain-name>
```

Destroy certificate

In order to destroy a service, it must not be in use by any load balancers or
any other AWS resources.

[region-table]: https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/
[go-sdk]: https://aws.amazon.com/documentation/sdk-for-go/
[go-env-vars]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#environment-variables
[go-shared-credentials-file]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#shared-credentials-file
[go-iam-roles-for-ec2-instances]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#iam-roles-for-ec2-instances
[go-specifying-credentials]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
[cwl-filter-expression]: http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-events
[acm-import-cert]: http://docs.aws.amazon.com/acm/latest/APIReference/API_ImportCertificate.html 
