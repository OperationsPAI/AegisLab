PWD := $(shell pwd)

builder:
	docker run -itd \
		-v /var/lib/dagger \
		--name dagger-engine-v0.14.0 \
		--privileged \
		-v $(PWD)/manifests/engine.toml:/etc/dagger/engine.toml \
		registry.dagger.io/engine:v0.14.0