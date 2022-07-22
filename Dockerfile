#v1.17 go
FROM golang@sha256:bd9823cdad5700fb4abe983854488749421d5b4fc84154c30dae474100468b85 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build

# TODO: use distroless:
# FROM gcr.io/distroless/base:nonroot@sha256:02f667185ccf78dbaaf79376b6904aea6d832638e1314387c2c2932f217ac5cb
FROM debian:11.4-slim@sha256:f576b8067b77ff85c70725c976b7b6cde960898e2f19b9abab3fb148407614e2

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    # For debugging.
    jq ca-certificates curl
COPY --from=build /src/scorecard-action /

# Copy a test policy for local testing.
COPY policies/template.yml  /policy.yml

ENTRYPOINT [ "/scorecard-action" ]
