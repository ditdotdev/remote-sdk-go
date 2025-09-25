# About this Project

This project provides a common set of interfaces for remote providers (such
as SSH and S3). Titan does not yet provide a way to dynamically load
remote providers, but this SDK provides a means for providers to have their
own dedicated repositories.

## CI/CD Pipeline

This repository now includes a comprehensive Pull Request 2 workflow with:
- Cross-platform testing (Ubuntu, Windows, macOS)
- Multi-version Go support (1.21, 1.22, 1.23)
- Security scanning and code quality checks
- Coverage reporting and performance benchmarks

## Contributing

This project follows the Titan community best practices:

  * [Contributing](https://github.com/titan-data/.github/blob/master/CONTRIBUTING.md)
  * [Code of Conduct](https://github.com/titan-data/.github/blob/master/CODE_OF_CONDUCT.md)
  * [Community Support](https://github.com/titan-data/.github/blob/master/SUPPORT.md)

It is maintained by the [Titan community maintainers](https://github.com/titan-data/.github/blob/master/MAINTAINERS.md)

For more information on how it works, and how to build and release new versions,
see the [Development Guidelines](DEVELOPING.md).

## License

This is code is licensed under the Apache License 2.0. Full license is
available [here](./LICENSE).
