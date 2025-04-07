FROM gcr.io/distroless/static:nonroot

# `nonroot` coming from distroless
USER 65532:65532

# Copy the binary that was built
COPY core /bin/core

# Change permissions to make it executable
RUN chmod +x /bin/core

# Copy default model into image
COPY fga/model/model.fga fga/model/model.fga

# Run the web service on container startup.
ENTRYPOINT [ "/bin/core" ]
CMD ["serve"]
