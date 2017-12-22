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
- fargate service create APPNAME
- fargate service deploy APPNAME
- fargate service info APPNAME
- fargate service logs APPNAME
- fargate service ps APPNAME
- fargate service scale APPNAME
- fargate service env set APPNAME
- fargate service env unset APPNAME
- fargate service env list APPNAME
- fargate service update APPNAME
- fargate service restart APPNAME
- fargate service destory APPNAME

### Tasks

- fargate task list
- fargate task start TASKNAME
- fargate task stop TASKNAME
- fargate task logs TASKNAME
- fargate task ps TASKNAME
- fargate task info TASKNAME

### Certificates

- fargate certificate list
- fargate certificate import CERTNAME
- fargate certificate request CERTNAME
- fargate certificate info CERTNAME
- fargate certificate validate CERTNAME
- fargate certificate destroy CERTNAME

## Load Balancers

- fargate lb list
- fargate lb create LBNAME
- fargate lb destroy LBNAME
- fargate lb alias <load balancer name> <hostname>
