FROM golang:1.12.6 as build-env
# All these steps will be cached
RUN mkdir /test
WORKDIR /test
#<- COPY go.mod and go.sum files to the workspace
COPY go.mod . 
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./main.go
# <- Second step to build minimal image
FROM scratch 
COPY --from=build-env /test/main /app/
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app/main"]

EXPOSE 80