## 0.3.0 (In Progress)

### Bug Fixes

- service create will not run if a load balancer is configured without a port.
- service create and task run will no longer create a repository if an image is
  explictly passed.
- service destroy will remove all references the service's target group and
  delete it

## 0.2.0 (2017-12-31)

### Features

- Added **--cluster** global flag to allow running commands against other
  clusters rather than the default. If omitted, the default **fargate** cluster
  is used. ([#2][issue-2])
- lb create, service create, and task run now accept an optional **--subnet-id**
  flag to place resources in different VPCs and subnets rather than the
  defaults. If omitted, resources will be placed within the default subnets
  within the default VPC. ([#2][issue-2])
- lb create, service create, and task run now accept an optional
  **--security-group-id** flag to allow applying more restrictive security
  groups to load balancers, services, and tasks. This flag can be passed
  multiple times to apply multiple security groups. If omitted, a permissive
  security group will be applied.

### Bug Fixes

- Resolved crashes with certificates missing resource records. Certificates that
  fail to be issued immediately after request would cause crashes in lb info and
  lb list as the resource record was never generated.

[issue-2]: https://github.com/jpignata/fargate/issues/2
