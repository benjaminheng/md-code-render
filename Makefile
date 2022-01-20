all:
	go install ./...

.PHONY: example
example:
	md-code-renderer render --languages dot --output-dir example/ ./README.md
