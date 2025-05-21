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
### Environment variables
```
SP_CALL_ACCESS_TOKEN=
SP_CALL_URL=
SP_SMS_ACCESS_TOKEN=
SP_SMS_URL=
SP_PORT=
DEBUG=TRUE
```

## Run locally
### Local build:
1. Create `config.yaml` in root directory of the project
2. `go run .`

### Docker:

1. Create `.env.smseagle-proxy` in the `local_testing` directory
2. Run `docker compose up --build` - this will start Grafana with OnCall
3. Setup Grafana OnCall:
- `curl -X POST 'http://admin:admin@localhost:3000/api/plugins/grafana-oncall-app/settings' -H "Content-Type: application/json" -d '{"enabled":true, "jsonData":{"stackId":5, "orgId":100, "onCallApiUrl":"http://engine:8080/", "grafanaUrl":"http://localhost:3000/"}}'`
- `curl -X POST 'http://admin:admin@localhost:3000/api/plugins/grafana-oncall-app/resources/plugin/install'`
- Go to Grafana at http://localhost:3000, user:pass `admin:admin`
- Add `smseagle-proxy` as an outgoing webhook
4. To rebuild `smseagle-proxy` and get logs run `docker compose up --build smseagle_proxy`

## Sending requests
Other than an OnCall json we need also a header with the `phonenumber` key.