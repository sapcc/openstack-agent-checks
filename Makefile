SHELL := /bin/bash
all: dep
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir"; \
	done

dep:
	dep ensure

clean:
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir" clean; \
	done
