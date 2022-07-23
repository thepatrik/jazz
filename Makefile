.DEFAULT_GOAL:= jazz

jazz:
	cd gojazz && go run .

.PHONY: jazz