# drone-teams

Drone plugin to send teams notifications for build status

## Usage

```console
docker run --rm \
  -e PLUGIN_WEBHOOK=<WEBHOOK ENDPOINT> \
  uchugroup/drone-teams
```

## Drone Pipeline Usage

```yaml
- name: teams-webhook
  image: uchugroup/drone-teams
  settings:
    webhook: <WEBHOOK ENDPOINT>
    facts:
    - 
```

With custom facts:

```yaml
- name: teams-webhook
  image: uchugroup/drone-teams
  settings:
    webhook: <WEBHOOK ENDPOINT>
    facts:
    - "Stage:TEST"
```

With log for steps with error:

```yaml
- name: teams-webhook
  image: uchugroup/drone-teams
  settings:
    webhook: <WEBHOOK ENDPOINT>
    logs_on_error: true
    logs_auth_token: 
      from_secret: logs_auth_token
```

Logs require a drone-ci auth token with admin permissions to get the logs
Without admin permissions drone will give a not found error when getting the logs

You can use openssl to create a token and drone cli to create a machine user with the generated token and admin permissions
Use this token as you auth_token

```console
foo@bar:~$ openssl rand -hex 16
85951798d79fbdb2e3823ab2d35f6a69

foo@bar:~$ drone user add logsreader --machine --admin --token=85951798d79fbdb2e3823ab2d35f6a69
```

Success sample:

![Sample success](https://github.com/uchugroup/drone-teams/raw/master/sample_success.png)


Failure sample:

![Sample failure](https://github.com/uchugroup/drone-teams/raw/master/sample_failure.png)


Failure sample with logs:

![Sample failure with logs](https://github.com/uchugroup/drone-teams/raw/master/sample_failure_logs.png)


## Build

Build the binary with the following command:

```console
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GO111MODULE=on

go build -a -tags netgo -o release/linux/amd64/drone-teams ./cmd/drone-teams
```

## Docker

Build the Docker image with the following command:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file Dockerfile --tag uchugroup/drone-teams .
```

For arm64:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file Dockerfile.arm64 --tag uchugroup/drone-teams .
```

For armhf / armv7:

```console
docker build \
  --label org.label-schema.build-date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --label org.label-schema.vcs-ref=$(git rev-parse --short HEAD) \
  --file Dockerfile.armv7 --tag uchugroup/drone-teams .
```

