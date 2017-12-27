# Fargate CLI

![](doc/media/fargate.png)

Deploy serverless containers onto the cloud from your command line

## Usage

### Commands

#### Tasks

- fargate task list
- fargate task run TASKGROUPNAME [--count NUM] [--cpu UNITS] [--memory MiB] [--image DOCKERIMAGE] [--env KEY=VALUE]
- fargate task stop TASKGROUPNAME [--task TASKID]
- fargate task logs TASKGROUPNAME [--follow] [--start TIMEEXPRESSION] [--end TIMEEXPRESSION] [--filter FILTEREXPRESSION] [--task TASKID]
- fargate task ps TASKGROUPNAME
- fargate task info TASKGROUPNAME [--task TASKID]

#### Services

- fargate service list
- fargate service create SERVICENAME [--cpu UNITS] [--memory MiB] [--port PORTEXPRESSION] [--lb LBNAME] [--rule RULEEXPRESSION] [--image DOCKERIMAGE] [--env KEY=VALUE] [--num NUM]
- fargate service deploy SERVICENAME [--image DOCKERIMAGE]
- fargate service info SERVICENAME
- fargate service logs SERVICENAME [--follow] [--start TIMEEXPRESSION] [--end TIMEEXPRESSION] [--filter FILTEREXPRESSION] [--task TASKID]
- fargate service ps SERVICENAME
- fargate service scale SERVICENAME SCALEEXPRESSION
- fargate service env set SERVICENAME --env KEY=VALUE
- fargate service env unset SERVICENAME --key KEYNAME
- fargate service env list SERVICENAME
- fargate service update SERVICENAME [--cpu UNITS] [--memory MiB]
- fargate service restart SERVICENAME
- fargate service destroy SERVICENAME

#### Certificates

- fargate certificate list
- fargate certificate import DOMAINNAME --certificate FILE --key FILE [--chain FILE]
- fargate certificate request DOMAINNAME
- fargate certificate info DOMAINNAME
- fargate certificate validate DOMAINNAME
- fargate certificate destroy DOMAINNAME

### Load Balancers

- fargate lb list
- fargate lb create LBNAME --port PORTEXPRESSION [--certificate DOMAINNAME]
- fargate lb destroy LBNAME
- fargate lb alias LBNAME HOSTNAME
- fargate lb info LBNAME
