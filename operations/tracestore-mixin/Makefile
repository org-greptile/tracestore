.PHONY: all check install dashboards alerts rules fmt

all: install dashboards alerts rules

check: all
	git diff --exit-code -- ../tracestore-mixin-compiled/

install:
	jb install

dashboards:
	jsonnet -J vendor -S dashboards.jsonnet -m ../tracestore-mixin-compiled/dashboards/

alerts:
	jsonnet -J vendor -S alerts.jsonnet > ../tracestore-mixin-compiled/alerts.yaml

rules:
	jsonnet -J vendor -S rules.jsonnet > ../tracestore-mixin-compiled/rules.yaml

fmt:
	find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print | \
		xargs -n 1 -- jsonnetfmt -i
