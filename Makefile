# Copyright 2024 Hewlett Packard Enterprise Development LP
.DEFAULT_GOAL := terraschema

.PHONY: test
test: 
	go test ./...

.PHONY: lint
lint: 
	golangci-lint run

.PHONY: all
all: test lint

.PHONY: terraschema
terraschema: 
	@go build .

.PHONY: clean
clean:
	@rm -f terraschema

.PHONY: test-data
test-data:
	@for name in `ls test/modules`; do \
		go run . -i test/modules/$$name -o test/expected/$$name/schema.json --overwrite --allow-empty --ignore-variable "ignored" --ignore-variable "also_ignored"; \
		go run . -i test/modules/$$name -o test/expected/$$name/variables.json --overwrite --allow-empty --export-variables --ignore-variable "ignored" --ignore-variable "also_ignored"; \
		go run . -i test/modules/$$name -o test/expected/$$name/schema-disallow-additional.json --overwrite --allow-empty --disallow-additional-properties --ignore-variable "ignored" --ignore-variable "also_ignored"; \
		go run . -i test/modules/$$name -o test/expected/$$name/schema-nullable-all.json --overwrite --allow-empty --nullable-all --ignore-variable "ignored" --ignore-variable "also_ignored"; \
		go run . -i test/modules/$$name -o test/expected/$$name/schema-with-title.json --overwrite --allow-empty --root-property "title=Example Schema" --root-property '$$id=http://example.com/schema' --ignore-variable "ignored" --ignore-variable "also_ignored"; \
	done
