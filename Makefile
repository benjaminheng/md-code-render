all:
	go install ./...

.PHONY: readme
readme:
	md-code-renderer render --languages dot --output-dir example/ --link-prefix "./example/" ./README.md
