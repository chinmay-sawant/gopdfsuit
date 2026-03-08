# Release Checklist 5.0.0

This checklist covers the v5 major-version release work for GoPdfSuit, including the explicit Go module versioning requirements introduced by the `/v5` module path.

## Release Scope

- [ ] Confirm the release target is `v5.0.0` across Docker tags, Go module paths, docs, and release notes.
- [ ] Confirm the repository is clean and all intended release changes are committed.

## Go Module Versioning

- [ ] Verify the root module path is `github.com/chinmay-sawant/gopdfsuit/v5` in `go.mod`.
- [ ] Verify nested module declarations under `bkp`, `certs`, `dockerfolder`, `frontend`, `guides`, `sampledata`, `screenshots`, `temp_verapdf`, and `verapdf` also use `/v5`.
- [ ] Verify all maintained Go imports reference `/v5` instead of `/v4`.
- [ ] Verify `sampledata/go.mod` requires `github.com/chinmay-sawant/gopdfsuit/v5 v5.0.0` and its `replace` points to the local repo.
- [ ] Run `go mod tidy` where needed after the module-path migration.

## Docker Release

- [ ] Build the Docker image with the release tag: `docker build -f dockerfolder/Dockerfile --build-arg VERSION=5.0.0 -t gopdfsuit:5.0.0 .`
- [ ] Smoke-test the image locally: `docker run --rm -p 8080:8080 gopdfsuit:5.0.0`.
- [ ] Tag the Docker image for Docker Hub: `docker tag gopdfsuit:5.0.0 chinmaysawant/gopdfsuit:5.0.0`.
- [ ] Update the `latest` tag if this release should become the default: `docker tag gopdfsuit:5.0.0 chinmaysawant/gopdfsuit:latest`.
- [ ] Push both tags: `docker push chinmaysawant/gopdfsuit:5.0.0` and `docker push chinmaysawant/gopdfsuit:latest`.
- [ ] Verify the remote image digest and pullability after push.

## Frontend And Documentation

- [ ] Verify frontend documentation uses explicit install instructions: `go get github.com/chinmay-sawant/gopdfsuit/v5@v5.0.0`.
- [ ] Verify all public code examples import `github.com/chinmay-sawant/gopdfsuit/v5/pkg/gopdflib`.
- [ ] Rebuild the frontend so the checked-in docs bundle under `docs/` reflects the updated v5 examples.
- [ ] Spot-check the documentation page in the built site after regeneration.

## Python Binding Consistency

- [ ] Verify `bindings/python/pyproject.toml` is `5.0.0`.
- [ ] Verify `bindings/python/pypdfsuit/__init__.py` reports `__version__ = "5.0.0"`.
- [ ] Rebuild Python binding artifacts so generated files and headers are aligned with the v5 Go module path.
- [ ] Recreate package metadata if a source distribution or wheel will be published.

## Validation

- [ ] Run Go tests: `go test ./...`.
- [ ] Run integration tests: `go test -count=1 -v ./test`.
- [ ] Run frontend checks from `frontend/`: `npm install` if needed, then `npm run build` and `npm run lint`.
- [ ] Run Python tests if publishing bindings: `python3 -m pytest bindings/python/tests`.
- [ ] Verify at least one `go get github.com/chinmay-sawant/gopdfsuit/v5@v5.0.0` consumer flow from a clean module.

## Release Publishing

- [ ] Prepare release notes summarizing the major-version migration from `/v4` to `/v5`.
- [ ] Create and push the git tag: `git tag v5.0.0` and `git push origin v5.0.0`.
- [ ] Publish the GitHub release with upgrade guidance for existing `/v4` consumers.
- [ ] Call out the breaking change clearly: imports must move from `/v4` to `/v5` and installs must use an explicit `@v5.0.0` tag.

## Post-Release Verification

- [ ] Pull the published Docker image on a clean machine or CI runner and verify startup.
- [ ] Verify the GitHub release, tag, and docs all point to `v5.0.0`.
- [ ] Confirm there are no remaining maintained `/v4` or `v4.2.0` references in source/docs.