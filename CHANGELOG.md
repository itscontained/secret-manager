# Changelog

<!---
All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
-->

## [Unreleased](https://github.com/itscontained/secret-manager/compare/v0.3.0...HEAD)

## [0.3.0](https://github.com/itscontained/secret-manager/compare/v0.2.0...v0.3.0)
- Fix double base64 encoding of secrets ([#59](https://github.com/itscontained/secret-manager/pull/59) [@devth](https://github.com/devth))
- Fix incorrect service account token path for vault store authentication ([#66](https://github.com/itscontained/secret-manager/pull/66) [@huguesalary](https://github.com/huguesalary))
- Fix nil pointer panic on an error during vault store authentication([#65](https://github.com/itscontained/secret-manager/pull/65) [@huguesalary](https://github.com/huguesalary))
- Fix Vault API path for v1 secret engine ([#42](https://github.com/itscontained/secret-manager/pull/42) [@c35sys](https://github.com/c35sys))
- Add E2E testing structure and tests for AWS Secret Manager ([#39](https://github.com/itscontained/secret-manager/pull/39) [@moolen](https://github.com/moolen))
- Fix logging flag registration ([#46](https://github.com/itscontained/secret-manager/pull/46) [@mcavoyk](https://github.com/mcavoyk))
- Update helm chart to install CRD's by default ([#68](https://github.com/itscontained/secret-manager/pull/68) [@mcavoyk](https://github.com/mcavoyk))
- Change base docker image from `gcr.io/distroless/static` to `alpine:3.12` ([#67](https://github.com/itscontained/secret-manager/pull/67) [@mcavoyk](https://github.com/mcavoyk))

## [0.2.0](https://github.com/itscontained/secret-manager/compare/v0.1.0...v0.2.0) - 2020-09-17
- Add GCP Secret Manager store backend ([#36](https://github.com/itscontained/secret-manager/pull/36) [@DirtyCajunRice](https://github.com/DirtyCajunRice))
