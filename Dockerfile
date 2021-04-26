FROM golang:1.16 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod verify

ARG version="DEV"

COPY *go .

RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s -X main.version=${version}" -o main .

############################
# STEP 2 build service image
############################

FROM scratch

ARG COMMIT_SHA=<not-specified>

WORKDIR /app

COPY --from=builder /app/* ./

# Use an unprivileged user.
USER 1000

CMD ["/app/main"]
