FROM cgr.dev/chainguard/bash:latest

# `nonroot` coming from distroless
USER 65532:65532

# Copy the binary that was built
COPY core /bin/core

# Copy default model into image
COPY fga/model/ fga/model/

# Run the web service on container startup.
ENTRYPOINT [ "/bin/core" ]
CMD ["serve"]
