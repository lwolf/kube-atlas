
.PHONY: build
build:
	bazel build //...

.PHONY: binary
binary:
	bazel build //:kube-atlas

.PHONY: watch
watch:
	ibazel build //:kube-atlas

.PHONY: gazelle
gazelle:
	bazel run //:gazelle

.PHONY: deps
deps:
	go mod tidy
	bazel run //:gazelle -- update-repos -from_file=go.mod
	@make gazelle

.PHONY: run
run:
	bazel run //:kube-atlas

.PHONY: run-binary
run-binary:
	./bazel-bin/darwin_amd64_stripped/kube-atlas

.PHONY: test
test:
	bazel test //...

.PHONY: clean
clean:
	bazel clean --expunge

.PHONY: push
push:
	bazel run //cmd/app:push
