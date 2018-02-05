## 0.3.0 (Unreleased)

### Enhancements

- Console output reworked for consistency and brevity
- macOS users get emoji as a type prefix in console output :tada: -- disable
  with --no-emoji if you're not into fun
- Requests and responses from AWS are displayed in full when --verbose is
  passed

### Bug Fixes

- Environment variable service commands now return a polite error message when
  invoked without the service name. ([#22][issue-22])
- Certificate import command re-implemented to work correctly. Previously calls
  to this command always returned "private key not supported" as we were
  incorrectly encoding it to base64 before passing it to the AWS SDK.

### Chores

- Utilize `dep` for dependency management
- Add contributor guide, updated license to repo

## 0.2.3 (2018-01-19)

### Features

- Support **--task-role** flag in service create and task run to allow passing
  a role name for the tasks to assume. ([#8][issue-8])

### Enhancements

- Use the `ForceNewDeployment` feature of `UpdateService` in service restart
  instead of incrementing the task definition. ([#14][issue-14])

### Bug Fixes

- Fixed issue where we'd stomp on an existing task role on service updates like
  deployments or environment variable changes. ([#8][issue-8])

## 0.2.2 (2018-01-11)

### Bug Fixes

- Fix service update operation to properly validate and run. ([#11][issue-11])
- Bail out early in service info if the requested service is not active meaning
  it has been previously destroyed.

## 0.2.1 (2018-01-02)

### Bug Fixes

- service create will not run if a load balancer is configured without a port.
- service create and task run will no longer create a repository if an image is
  explictly passed.
- service destroy will remove all references the service's target group and
  delete it.
- Fix git repo detection to properly use a git sha image tag rather than a
  time stamp tag. ([#6][issue-6])
- Fail fast if a user attempts to destroy a service scaled above 0.

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
[issue-6]: https://github.com/jpignata/fargate/issues/6
[issue-8]: https://github.com/jpignata/fargate/issues/8
[issue-11]: https://github.com/jpignata/fargate/issues/11
[issue-14]: https://github.com/jpignata/fargate/issues/14
[issue-22]: https://github.com/jpignata/fargate/issues/22
