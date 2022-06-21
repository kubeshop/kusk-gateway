# Smoketest

## Running smoketest:

Ensure you have a cluster running and have Kusk Gateway and an EnvoyFleet deployed.

For minikube, you can do this with make `create-env`

Run all smoke tests
```
make check-all
```

or individually:

```
make check-basic 
```

## Adding smoketests

1. Before adding a smoke test first add test in Makefile.variables in form check-{test_name}
2. Next add a directory in smoketest {test_name}. 
3. Ensure that you are implementing tests like this:

```
type SomeTestCheckSuite struct {
	common.KuskTestSuite
}
```
