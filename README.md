# URL Visitor test exercise

## Usage
There are two ways to provide URLs to the app
* through STDIN 
* via cli arguments

### Via Makefile
```bash
make run-urls URLS="example.com ya.ru httpbin.com"
```
this is equal to `./bin/app example.com ya.ru httpbin.com`
or
```bash
make run-file
```
this is equal to `cat ./urls.txt | ./bin/app` 

Concurrency can be controlled via ENV var `MAX_CONCURRENCY`
HTTP Timeout can be controlled via ENV var `HTTP_TIMEOUT_SECONDS`
