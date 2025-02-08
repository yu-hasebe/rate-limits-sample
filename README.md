# rate-limits-sample

This is a rate-limits sample composed of Go and Redis.

```
$ docker-compose up

# You can call this API so many times that you will see the response with status code 429.
$ curl localhost:8080
```
