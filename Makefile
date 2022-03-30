
test:
	go test ./...
test_os:
	go test -v driver/os
test_mem:
	go test -v driver/memory

.PHONY: test test_os test_mem