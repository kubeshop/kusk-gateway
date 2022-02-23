# Integration testing

This is the minimal setup to do the simple integration testing:

- All resources are located in **testing** namespace, including new EnvoyFleet testing.testing. This makes it easy to delete all testing artifacts.
- Single TodoMVC backend and frontend.
- TodoMVC backend and frontend routing (API and StaticRoute) are setup with /testing prefix to differentiate with possible present usual default ToDoMVC during the local development.
- example.com and example.org Hosts are used to route to TodoMVC backend and frontend.
- There is the second TodoMVC API object available only for example.org/second that has different x-kusk options - either redirects to or is using (with the rewrite) the TodoMVC only for GET requests for /todos. Other endpoints are disabled.
- There is the second TodoMVC StaticRoute object available only for example.org/testing/staticroute/second to test the different path specific only for that hostname.
- Postman requests collections are run using Newman to test HTTP calls and the responses.

## Known Bugs

- Current TodoMVC frontend can't work correctly with other than default base url, so deleting and editing of the items from the Frontend doesn't work.

## How To develop tests

1. Write the test case in this README - the clear description what we're testing.

2. Import the existing Postman collections from "postman" directory into Postman Desktop app, add the related test where applicable. Note that we test against EXTERNAL_IP (load balancer address) and we need to add the related Host header to the request therefore.

3. After that export the collection and put it into "postman" directory.

4. Run "runtest.sh delete && runtest.sh all" to verify that everything works.

## ToDO MVC testcases

### API testing

1. Prefix and rewrite

    - API is available on example.com/testing prefix - returns json object and 200.
    - POST to example.com/testing/todos (with the prepared body) works, returns 201.
    - The subsequent GET to example.com/testing/todos/1 (with the prepared body) works, returns 200 and the same json object body.

2. CORS, with overriding

    - Correct CORS headers are available on http://example.com/testing/todos with GET and OPTIONS when Origin: example.com.
    - Other CORS headers are available on http://example.com/testing/todos/1 with GET and OPTIONS when Origin: example.com.

3. Hosts

    - /testing/todos are available on example.com AND example.org but not any other Host (example.net) - example.net should return 404.

4. Disabled

    On second:

    - GET http://example.org/second/todos is present, but POST http://example.org/second/todos fails with 404.

5. Redirect

    On second:

    - Enabled example.org/second/todos/{id} GET to redirect to example.com/testing/todos/{id}. This tests the regex as well.

6. Mocking

   Mocking is done using Agent sidecar.
    - Enabled example.com/testing/mocking/{id} GET to respond with mocked payload from "example" OpenAPI field.
    - Enabled example.com/testing/mocking/{id} DELETE to respond with mocked HTTP status code, without the body.
    - Enabled example.com/testing/mocking/multiple/{id} GET to respond with mocked payload from "examples" OpenAPI field.
    - Disabled example.com/testing/mocking/multiple/{id} PATCH to respond with mocked payload from "examples" OpenAPI field.

### StaticRoute

1. Prefix and Rewrite

    - Front is available on example.com/testing/ prefix - returns text object and 200.
    - API is available on example.com/testing/staticroute/ prefix - returns json object and 200.
    - POST to example.com/testing/staticroute/todos (with the prepared body) works, returns 201.
    - The subsequent GET to example.com/testing/staticroute/todos/{id} (with the prepared body) works, returns 200 and the same json object body as previously submitted.

2. CORS

    - Correct CORS headers are available on http://example.com/testing/staticroute/todos with GET and OPTIONS when Origin: example.com.
    - Other specified correct CORS headers are available on http://example.com/testing/staticroute/todos/1 with GET and OPTIONS when Origin: example.com.

3. Hosts

    - /testing/staticroute/todos are available on example.com AND example.org but not any other Host (example.net) - example.net should return 404.
    - /second/staticroute/todos are available only on example.org and NOT example.com, should return 404.

4. Redirect

    - GET to http://example.com/ - redirected to http://example.com/testing/ with StaticRoute, response code 301.
    - GET to http://example.com/non-existent - redirected to http://example.com/testing/non-existent with StaticRoute, response code 308.

### TODO items

- Websockets testing is currently hard to script, so for now this is manual testing using /examples/websocket directory.

- QoS (Retries, Timeouts)
