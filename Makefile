statistics_lines:
	@core_lines=$$(find . -name '*.go' ! -name '*_test.go' -exec wc -l {} + | awk '{ last = $$1 } END { print last + 0 }'); \
	total_lines=$$(find . -name '*.go' -exec wc -l {} + | awk '{ last = $$1 } END { print last + 0 }'); \
	echo "Core Lines: $$core_lines"; \
	echo "Total Lines: $$total_lines"

unittest:
	mkdir -p .coverage; \
	report=$$(pwd)/unittest_coverage.md; \
	echo "# Unit Test Coverage" > $$report; \
	echo "" >> $$report; \
	echo "Generated at: $$(date '+%Y-%m-%d %H:%M:%S %z')" >> $$report; \
	echo "" >> $$report; \
	echo "| Module | Coverage | Profile |" >> $$report; \
	echo "| --- | --- | --- |" >> $$report; \
	echo "mode: count" > .coverage/all.cover.out; \
	for module in $$(find . -maxdepth 2 -name go.mod | sort | xargs -n1 dirname); do \
		echo "==> $$module"; \
		if [ "$$module" = "." ]; then \
			echo "skip $$module (no packages)"; \
		elif [ -n "$$(find $$module -name '*.go' -print -quit)" ]; then \
			module_name=$$(basename $$module); \
			profile=$$(pwd)/.coverage/$$module_name.cover.out; \
			(cd $$module && go test -covermode=count -coverprofile=$$profile ./...) || exit 1; \
			coverage=$$(awk 'BEGIN { total=0; covered=0 } /^mode:/ { next } { total+=$$2; if ($$3 > 0) covered+=$$2 } END { if (total == 0) printf "0.0%%"; else printf "%.1f%%", covered/total*100 }' $$profile); \
			echo "coverage total: $$coverage"; \
			tail -n +2 $$profile >> .coverage/all.cover.out; \
			echo "| $$module_name | $$coverage | \`.coverage/$$module_name.cover.out\` |" >> $$report; \
			echo "coverage profile: $$profile"; \
		else \
			echo "skip $$module (no packages)"; \
		fi; \
	done; \
	echo "==> overall"; \
	overall_coverage=$$(awk 'BEGIN { total=0; covered=0 } /^mode:/ { next } { total+=$$2; if ($$3 > 0) covered+=$$2 } END { if (total == 0) printf "0.0%%"; else printf "%.1f%%", covered/total*100 }' .coverage/all.cover.out); \
	echo "project coverage total: $$overall_coverage"; \
	echo "" >> $$report; \
	echo "## Overall" >> $$report; \
	echo "" >> $$report; \
	echo "- Project coverage total: $$overall_coverage" >> $$report
