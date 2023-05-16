run: build
	./example-saml-service-provider

build: example-saml-service-provider

example-saml-service-provider: main.go example-saml-service-provider-key.pem
	go build -o $@

example-saml-service-provider-key.pem:
	openssl req \
		-x509 \
		-newkey rsa:2048 \
		-sha256 \
		-keyout $@ \
		-out $(@:-key.pem=-crt.pem) \
		-days 365 \
		-nodes \
		-subj //CN=localhost

clean:
	rm -f example-saml-service-provider example-saml-service-provider-metadata.xml *.pem

.PHONY: build run clean
