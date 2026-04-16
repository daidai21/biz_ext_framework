statistics_lines:
	find . | grep "\.go" | grep -v "_test\.go" | xargs wc -l

unittest:
	for module in $$(find . -maxdepth 2 -name go.mod | sort | xargs -n1 dirname); do \
		echo "==> $$module"; \
		packages=$$(cd $$module && go list ./... 2>/dev/null); \
		if [ -n "$$packages" ]; then \
			(cd $$module && go test ./...) || exit 1; \
		else \
			echo "skip $$module (no packages)"; \
		fi; \
	done
