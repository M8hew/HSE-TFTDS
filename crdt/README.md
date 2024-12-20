# Operation-based conflict-free replicated map

## Examples of Curl queries for testing

```sh
curl -X GET http://localhost:8080/state
```

```sh
curl -X POST http://localhost:8080/switch
```

```sh
curl -X PATCH http://localhost:8080/update \
-H "Content-Type: application/json" \
-d '{
    "operation":"add",
    "key":"key1",
    "value":"value1"
    }'
```

```sh
curl -X PATCH http://localhost:8080/update \
-H "Content-Type: application/json" \
-d '{
    "operation":"del",
    "key":"key2",
    "value":"value2"
    }'
```