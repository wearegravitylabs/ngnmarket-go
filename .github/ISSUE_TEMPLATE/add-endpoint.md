---
name: Add an endpoint
about: Wrap one NGN Market API endpoint in the SDK
title: "Add GET /<route> (<MethodName>)"
labels: endpoint
---

## Endpoint

- **Route:** `GET /...`
- **Suggested method name:** `...`
- **Docs:** https://docs.ngnmarket.com/...
- **Lowest plan that can call it:** Free / Hobby / Starter / Pro / Business
- **Paged?** yes / no (if yes: `page` and `limit`, and the response has a `pagination` object)

## What to do

Follow the recipe in [TASKS.md](../../TASKS.md#3-the-recipe-add-one-endpoint) (section 3).

- [ ] Types in `model/` (and a `...Params` struct with `Values()` if it takes query parameters)
- [ ] Method added to the `RemoteCalls` interface in `api/api.go`
- [ ] Implementation in `api/` that calls `c.makeRequest`
- [ ] Unit test with a fake server and a real or docs response as the fixture
- [ ] `go generate ./...` run (refreshes `api/mock`)
- [ ] `./quality.sh` passes
- [ ] Row added to the README endpoint table (say "unverified" if it needs a paid plan you do not have)
- [ ] Line added to `CHANGELOG.md`
- [ ] No API key anywhere in the diff

## Notes

Anything useful: odd fields, `null` values, things you were unsure about.
