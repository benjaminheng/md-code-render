all:
	go install ./...

.PHONY: example
example:
	md-code-renderer render --languages dot --output-dir example/ --link-prefix "./example/" ./README.md
