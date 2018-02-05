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

- Ensure you're using golang 1.9

  ```console
  go version
  ```

- Install [`dep`][dep] if not present on your system. See their [installation
  instructions][dep-install] and [releases page][dep-releases] for details. You
  can also install the latest through `go install`

  ```console
  go install github.com/golang/dep
  ```

- Install the source code from GitHub

  ```console
  go get github.com/jpignata/fargate
  ```

- Run `dep ensure` to install required dependencies

  ```console
  cd $GOPATH/src/github.com/jpignata/fargate
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

## Licensing

This project is released under an [Apache 2.0][apache] license.

## Code of Conduct

This project abides by the [Amazon Open Source Code of Conduct][amzn-coc].
Please be nice.

### Amazon Open Source Code of Conduct

This code of conduct provides guidance on participation in Amazon-managed open
source communities, and outlines the process for reporting unacceptable
behavior. As an organization and community, we are committed to providing an
inclusive environment for everyone. Anyone violating this code of conduct may be
removed and banned from the community.

**Our open source communities endeavor to:**
- Use welcoming and inclusive language;
- Be respectful of differing viewpoints at all times;
- Accept constructive criticism and work together toward decisions;
- Focus on what is best for the community and users.

**Our Responsibility.** As contributors, members, or bystanders we each
individually have the responsibility to behave professionally and respectfully
at all times. Disrespectful and unacceptable behaviors include, but are not
limited to: The use of violent threats, abusive, discriminatory, or derogatory
language;
- Offensive comments related to gender, gender identity and expression, sexual
  orientation, disability, mental illness, race, political or religious
  affiliation;
- Posting of sexually explicit or violent content;
- The use of sexualized language and unwelcome sexual attention or advances;
- Public or private
  [harassment](http://todogroup.org/opencodeofconduct/#definitions) of any kind;
- Publishing private information, such as physical or electronic address,
  without permission;
- Other conduct which could reasonably be considered inappropriate in a
  professional setting;
- Advocating for or encouraging any of the above behaviors.

**Enforcement and Reporting Code of Conduct Issues.** Instances of abusive,
harassing, or otherwise unacceptable behavior may be reported by contacting
opensource-codeofconduct@amazon.com. All complaints will be reviewed and
investigated and will result in a response that is deemed necessary and
appropriate to the circumstances.

**Attribution.** _This code of conduct is based on the
[template](http://todogroup.org/opencodeofconduct) established by the [TODO
Group](http://todogroup.org/) and the Scope section from the [Contributor
Covenant version 1.4](http://contributor-covenant.org/version/1/4/)._

[dep]: https://golang.github.io/dep
[dep-install]: https://golang.github.io/dep/docs/installation.html
[dep-releases]: https://github.com/golang/dep/releases
[amzn-coc]: https://aws.github.io/code-of-conduct
[apache]: http://aws.amazon.com/apache-2-0/
