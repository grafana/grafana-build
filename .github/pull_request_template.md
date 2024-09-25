# Requirements

The `main` branch of `grafana-build` should be compatible with all active versions of Grafana **and Grafana-Enterprise**. To
run integration tests, add a comment to this PR with the following:

```
/grafana-integration-tests
```

* [ ] I have ran the integraiton tests `gh workflow run --ref=${THIS_BRANCH} --repo=grafana/grafana-build pr-integration-tests.yml`
* [ ] All integration tests have passed

