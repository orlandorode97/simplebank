# Build stage

# Pulling official golang image as builder stage
FROM golang:1.20-rc-alpine as builder 

## Install make command to build the project and other tools.
RUN apk add g++ make wget gcc
RUN go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

## Create a working directory for the container
WORKDIR /app 
RUN wget https://github.com/pressly/goose/releases/download/v3.7.0/goose_linux_x86_64
RUN wget https://github.com/eficode/wait-for/releases/download/v2.2.3/wait-for

## Copy all from the root of the project to the root of the working directory (/app)
COPY . .
## Execute make build command to build the project
RUN make build

#Run stage

## Pulling alpine official image
FROM alpine:3.14

## Create a working directory for the second stage
WORKDIR /app
## Copy /bin/simplebank executable file from the builder stage to the current working
## of the second stage and rest of the files
COPY --from=builder /app/bin/simplebank .
COPY --from=builder /app/goose_linux_x86_64 .
COPY --from=builder /app/wait-for .
RUN ["chmod", "+x", "wait-for", "goose_linux_x86_64"]
COPY sql/migrations ./migrations
COPY app.env .
COPY start.sh .

## Expose to the outside world
EXPOSE 8081
## Execute bin/simplebank and script.
ENTRYPOINT [ "./start.sh", "./simplebank" ]

