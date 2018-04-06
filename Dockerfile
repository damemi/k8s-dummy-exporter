FROM gcr.io/google-containers/debian-base-amd64:0.1
COPY k8s_dummy_exporter /
USER 1001:1001
ENTRYPOINT ["/k8s_dummy_exporter"]