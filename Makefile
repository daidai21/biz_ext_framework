statistics_lines:
	find . | grep "\.go" | grep -v "_test\.go" | xargs wc -l

unittest:
	for module in $$(find . -maxdepth 2 -name go.mod | sort | xargs -n1 dirname); do \
		echo "==> $$module"; \
		if [ "$$module" = "." ]; then \
			echo "skip $$module (no packages)"; \
		elif [ -n "$$(find $$module -name '*.go' -print -quit)" ]; then \
			(cd $$module && go test ./...) || exit 1; \
		else \
			echo "skip $$module (no packages)"; \
		fi; \
	done
