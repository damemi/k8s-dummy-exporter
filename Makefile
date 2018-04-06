TAG = v0.1.0
PREFIX = docker.io/mdame

build: k8s_dummy_exporter

k8s_dummy_exporter: k8s_dummy_exporter.go
	go build -a -o k8s_dummy_exporter k8s_dummy_exporter.go

docker: k8s_dummy_exporter
	docker build -t ${PREFIX}/k8s-dummy-exporter:$(TAG) .

push: docker
	docker push ${PREFIX}/k8s-dummy-exporter:$(TAG)

clean:
	rm -rf k8s_dummy_exporter
