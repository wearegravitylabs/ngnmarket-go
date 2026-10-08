# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Initial SDK: `api` (client and `RemoteCalls` interface) and `model` (types and errors) packages covering NGX and
  US ticker listing/search, symbol lists, company/ticker detail and daily price history, plus
  `Get...ClosePriceOnOrBefore` helpers.
- Typed `model.APIError` with `errors.Is` sentinels, retries with backoff, optional client-side rate limiter.
- Generated gomock (`api/mock`), runnable samples in `examples/`, `quality.sh` and `.golangci.yml`.
