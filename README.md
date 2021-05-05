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

Success sample:

![Sample success](https://github.com/uchugroup/drone-teams/raw/master/sample_success.png)


Failure sample:

![Sample success](https://github.com/uchugroup/drone-teams/raw/master/sample_failure.png)


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

