# smseagle-proxy
A service to receive requests from Grafana Oncall and proxying them to SMSEagle.

## Config
### config.yaml
```yaml
call:
  access-token: 
  url: 
sms:
  access-token: 
  url:  
debug: true
port: 8095//default
```
### env variables
```
SP_CALL_ACCESS_TOKEN=
SP_CALL_URL=
SP_SMS_ACCESS_TOKEN=
SP_SMS_URL=
SP_PORT=
DEBUG=TRUE
```

## Run locally
### local build:
```
1. create config.yaml in root directory of the project
2. go run .
```
### docker:
```
1. create .env.smseagle-proxy in the local_testing directory
2. run docker compose up --build - this will start grafana with oncall
3. setup grafana oncall:
    a. go to grafana at http://localhost:3000, user:pass admin:admin
    b. enable oncall plugin: Administration -> Plugins -> Search for oncall -> Oncall backend url: http://engine:8080
    c. add smseagle-proxy as an outgoing webhook
4. to rebuild smseagle-proxy and get logs run "docker compose up --build skyline"
```

### sending requests
other than an oncall json we need also a header with the `phonenumber` key.