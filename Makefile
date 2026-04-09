statistics_lines:
	find . | grep "\.go$" | grep -v "_test\.go" | xargs wc -l
