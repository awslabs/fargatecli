# Fargate CLI

## Status

Experimental; very much a work in progress.

## Objective

Dead simple deployment and management of highly available container-based
applications.

### How?

The Fargate CLI provides a set of commands to allow users to quickly deploy and
monitor applications on top of AWS Fargate. It abstracts the details of Amazon
ECS, task definitions, target groups, services, Amazon VPC, security groups, and
a host of other nouns so developers can focus on what matters: their
applications.

## Commands

### Services

- fargate service list
- fargate service create SERVICENAME [--cpu UNITS] [--memory MiB] [--port PORTEXPRESSION] [--lb LBNAME] [--rule RULEEXPRESSION] [--image DOCKERIMAGE] [--env KEY=VALUE]
- fargate service deploy SERVICENAME [--image DOCKERIMAGE]
- fargate service info SERVICENAME
- fargate service logs SERVICENAME [--follow] [--start TIMEEXPRESSION] [--end TIMEEXPRESSION] [--filter FILTEREXPRESSION]
- fargate service ps SERVICENAME
- fargate service scale SERVICENAME SCALEEXPRESSION
- fargate service env set SERVICENAME --env KEY=VALUE
- fargate service env unset SERVICENAME --key KEYNAME
- fargate service env list SERVICENAME
- fargate service update SERVICENAME [--cpu UNITS] [--memory MiB]
- fargate service restart SERVICENAME
- fargate service destroy SERVICENAME

### Tasks

- fargate task list
- fargate task run TASKNAME [--count NUM] [--cpu UNITS] [--memory MiB] [--image DOCKERIMAGE] [--env KEY=VALUE]
- fargate task stop TASKNAME [--task TASKID]
- fargate task logs TASKNAME [--follow] [--start TIMEEXPRESSION] [--end TIMEEXPRESSION] [--filter FILTEREXPRESSION]
- fargate task ps TASKNAME
- fargate task info TASKNAME [--task TASKID]

### Certificates

- fargate certificate list
- fargate certificate import DOMAINNAME --certificate FILE --key FILE [--chain FILE]
- fargate certificate request DOMAINNAME
- fargate certificate info DOMAINNAME
- fargate certificate validate DOMAINNAME
- fargate certificate destroy DOMAINNAME

## Load Balancers

- fargate lb list
- fargate lb create LBNAME --port PORTEXPRESSION [--certificate DOMAINNAME]
- fargate lb destroy LBNAME
- fargate lb alias LBNAME HOSTNAME
- fargate lb info LBNAME
