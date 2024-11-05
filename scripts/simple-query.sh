#!/bin/sh

curl -XPUT -d '{"dateOfBirth": "2021-10-01"}' 'http://localhost:8080/hello/apple' -w '%{http_code}\n'
curl -XPUT -d '{"dateOfBirth": "2021-11-01"}' 'http://localhost:8080/hello/pear' -w '%{http_code}\n'
curl -XPUT -d '{"dateOfBirth": "2021-01-05"}' 'http://localhost:8080/hello/orange' -w '%{http_code}\n'

curl -XGET 'http://localhost:8080/hello/apple'
curl -XGET 'http://localhost:8080/hello/pear'
curl -XGET 'http://localhost:8080/hello/orange'
