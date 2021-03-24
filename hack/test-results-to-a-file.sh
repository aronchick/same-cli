#!/bin/bash
/usr/local/go/bin/go test -timeout 30s -run ^TestProgramCreateSuite$  github.com/azure-octo/same-cli/test/integration > test_results.txt; sed -i 's/\\n/\n/g' test_results.txt