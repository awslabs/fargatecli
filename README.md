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
- fargate service create APPNAME [--cpu UNITS] [--memory MiB] [--port PORTEXPRESSION] [--lb LBNAME] [--rule RULEEXPRESSION] [--image DOCKERIMAGE] [--env KEY=VALUE]
- fargate service deploy APPNAME [--image DOCKERIMAGE]
- fargate service info APPNAME
- fargate service logs APPNAME [--tail] [--num LINES]
- fargate service ps APPNAME
- fargate service scale APPNAME SCALEEXPRESSION
- fargate service env set APPNAME --env KEY=VALUE
- fargate service env unset APPNAME --key KEYNAME
- fargate service env list APPNAME
- fargate service update APPNAME [--cpu UNITS] [--memory MiB]
- fargate service restart APPNAME
- fargate service destroy APPNAME

### Tasks

- fargate task list
- fargate task start TASKNAME [--count TASKS] [--cpu UNITS] [--memory GB] [--image DOCKERIMAGE] [--env KEY=VALUE]
- fargate task stop TASKNAME
- fargate task logs TASKNAME [--tail] [--num LINES]
- fargate task ps TASKNAME
- fargate task info TASKNAME

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
