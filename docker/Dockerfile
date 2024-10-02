FROM golang:1.23.2 AS builder

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /go/bin/core -a -ldflags '-linkmode external -extldflags "-static"' .

FROM gcr.io/distroless/static:nonroot

# `nonroot` coming from distroless
USER 65532:65532

# Copy the binary that goreleaser built
COPY --from=builder /go/bin/core /bin/core

# Copy default model into image
COPY fga/model/model.fga fga/model/model.fga

# Run the web service on container startup.
ENTRYPOINT [ "/bin/core" ]
CMD ["serve"]
