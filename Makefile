all:
	go install ./...

.PHONY: readme
readme:
	md-code-renderer render --languages dot --output-dir example/ --link-prefix "./example/" ./README.md
	md-code-renderer clean --image-dir ./example README.md
