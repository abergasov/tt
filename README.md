# Field Materials Backend Interview skeleton code

## Build & Run Server locally
```
make run

# for run on specific port
make port=8081 run
```

## Run a sample request against the server
```
curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize
```

Now in your browser, you can check one of the returned urls!
