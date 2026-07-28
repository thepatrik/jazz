.DEFAULT_GOAL:= jazz

jazz:
	cd gojazz && go run .

test-c:
	cd cjazz && make test

.PHONY: jazz test-c