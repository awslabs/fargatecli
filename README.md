# Fargate CLI

![](doc/media/fargate.png)

Deploy serverless containers onto the cloud from your command line

## Usage

### Configuration

#### Region

By default, fargate uses *us-east-1* as this is the single Region where AWS
Fargate is available. The CLI accepts a --region parameter for future use and
will honor *AWS_REGION* and *AWS_DEFAULT_REGION* environment settings. Note that
specifying a Region where all required services aren't available will return an
error.

See the [Region Table][region-table] for a breakdown of what services are
available in which Regions.

#### Credentials

fargate is built using the [AWS SDK for Go][go-sdk] which looks for credentials
in the following locations:

1. [Environment Variables][go-env-vars]

1. [Shared Credentials File][go-shared-credentials-file]

1. [EC2 Instance Profile][go-iam-roles-for-ec2-instances]

For more information see [Specifying Credentials][go-specifying-credentials] in
the AWS SDK for Go documentation.

### Commands

#### Tasks

Tasks are one-time executions of your container. Instances of your task are run
until you manually stop them either through AWS APIs, the AWS Management
Console, or `fargate task stop`, or until they are interrupted for any reason.

##### fargate task list

List running task groups

##### fargate task run \<task-group-name> [--num \<count>] [--cpu \<cpu-units>] [--memory \<MiB>] [--image \<docker-image>] [--env \<key=value>]

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

##### fargate task info \<task-group-name> [--task \<task-id>]

Inspect tasks

Shows extended information for each running task within a task group or for
specific tasks specified with the --task flag. Information includes environment
variables which could differ between tasks in a task group. To inspect multiple
specific tasks within a task group specific --task with a task ID multiple
times.

##### fargate task ps \<task-group-name>

List running tasks

##### fargate task logs \<task-group-name> [--follow] [--start \<time-expression>] [--end \<time-expression>] [--filter \<filter-expression>] [--task \<task-id>]

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

##### fargate task stop \<task-group-name> [--task \<task-id>]

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

##### fargate service list

List services

##### fargate service create \<service name> [--cpu \<cpu units>] [--memory \<MiB>] [--port \<port-expression>] [--lb \<load-balancer-name>] [--rule \<rule-expression>] [--image \<docker-image>] [--env \<key=value>] [--num \<count>]

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
* to match multiple characters and ? to match a single character.

Environment variables can be specified via the --env flag. Specify --env with a
key=value parameter multiple times to add multiple variables.

Specify the desired count of tasks the service should maintain by passing the
--num flag with a number. If you omit this flag, fargate will configure a
service with a desired number of tasks of 1.

##### fargate service deploy SERVICENAME [--image DOCKERIMAGE]

Deploy new image to service

The Docker container image to use in the service can be optionally specified
via the --image flag. If not specified, fargate will build a new Docker
container image from the current working directory and push it to Amazon ECR in
a repository named for the task group. If the current working directory is a
git repository, the container image will be tagged with the short ref of the
HEAD commit. If not, a timestamp in the format of YYYYMMDDHHMMSS will be used.

##### fargate service info SERVICENAME

Inspect service

Show extended information for a service including load balancer configuration,
active deployments, and environment variables.

Deployments show active versions of your service that are running. Multiple
deployments are shown if a service is transitioning due to a deployment or
update to configuration such a CPU, memory, or environment variables. 

##### fargate service logs \<service-name> [--follow] [--start \<time-expression>] [--end \<time-expression>] [--filter \<filter-expression>] [--task \<task-id>]

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

##### fargate service ps \<service-name>

List running tasks for a service

##### fargate service scale \<service-name> \<scale-expression>

Scale number of tasks in a service

Changes the number of desired tasks to be run in a service by the given scale
expression. A scale expression can either be an absolute number or a delta
specified with a sign such as +5 or -2.

##### fargate service env set \<service-name> --env \<key=value>

Set environment variables

At least one environment variable must be specified via the --env flag. Specify
--env with a key=value parameter multiple times to add multiple variables.

##### fargate service env unset \<service-name> --key \<key-name>

Unset environment variables

Unsets the environment variable specified via the --key flag. Specify --key with
a key name multiple times to unset multiple variables.

##### fargate service env list \<service-name>

Show environment variables

##### fargate service update \<service-name> [--cpu \<cpu-units>] [--memory \<MiB>]

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

##### fargate service restart \<service-name>

Restart service

Creates a new set of tasks for the service and stops the previous tasks. This
is useful if your service needs to reload data cached from an external source,
for example.

##### fargate service destroy \<service-name>

Destroy service

Deletes a service. In order to destroy a service, it must first be scaled to 0
running.

#### Certificates

- fargate certificate list
- fargate certificate import DOMAINNAME --certificate FILE --key FILE [--chain FILE]
- fargate certificate request DOMAINNAME
- fargate certificate info DOMAINNAME
- fargate certificate validate DOMAINNAME
- fargate certificate destroy DOMAINNAME

#### Load Balancers

- fargate lb list
- fargate lb create LBNAME --port PORTEXPRESSION [--certificate DOMAINNAME]
- fargate lb destroy LBNAME
- fargate lb alias LBNAME HOSTNAME
- fargate lb info LBNAME

[region-table]: https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/
[go-sdk]: https://aws.amazon.com/documentation/sdk-for-go/
[go-env-vars]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#environment-variables
[go-shared-credentials-file]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#shared-credentials-file
[go-iam-roles-for-ec2-instances]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#iam-roles-for-ec2-instances
[go-specifying-credentials]: http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
[cwl-filter-expression]: http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-events
