**Welcome!** Thank you for considering contributing to this project! If I can
help in anyway to get you going, please feel free to reach out. I'm available by
email and Google Hangouts at john@pignata.com.

# Contributing

## Workflow

- **Did you find a bug?**

  Awesome! Please feel free to open an issue first, or if you have a fix open a
  pull request that describes the bug with code that demonstrates the bug in a
  test and addresses it.

- **Do you want to add a feature?**

  Features begin life as a proposal. Please open a pull request with a proposal
  that explains the feature, its use case, considerations, and design. This will
  allow interested contributors to weigh in, refine the idea, and ensure there's
  no wasted time in the event a feature doesn't fit with our direction.

## Setup

- Ensure you're using golang 1.9+.

  ```console
  go version
  ```

- Install [`dep`][dep] if not present on your system. See their [installation
  instructions][dep-install] and [releases page][dep-releases] for details.

- Install the source code from GitHub

  ```console
  go get github.com/awslabs/fargatecli
  ```

- Run `dep ensure` to install required dependencies

  ```console
  cd $GOPATH/src/github.com/awslabs/fargatecli
  dep ensure
  ```

- Make sure you can run the tests

  ```console
  make test
  ```

## Testing

- Tests can be run via `go test` or `make test`

- To generate mocks as you add functionality, run `make mocks` or use `go
  generate` directly

## Building

- To build a binary for your platform run `make`

- For cross-building for all supported platforms, run `make dist` which builds
  binaries for darwin (64-bit) and linux (Arm, 32-bit, 64-bit).

## Developing using Docker

### Minimum requirement

1. Visual Studio Code
2. Remote - Containers extension for Visual Studio Code
3. Docker

Clone the current repository.  Open the folder with VSCode Remote - Containers extension.

Note: The remote container has *AWS_SDK_LOAD_CONFIG* environment variable set to use AWS CLI configuration.

Open a terminal in VS Code to execute the cli.

```console
go run main.go --version
```


## Licensing

This project is released under the [Apache 2.0 license][apache].

## Code of Conduct

This project abides by the [Amazon Open Source Code of Conduct][amzn-coc].
Please be nice.

[dep]: https://golang.github.io/dep
[dep-install]: https://golang.github.io/dep/docs/installation.html
[dep-releases]: https://github.com/golang/dep/releases
[amzn-coc]: https://aws.github.io/code-of-conduct
[apache]: http://aws.amazon.com/apache-2-0/
