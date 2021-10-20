SHELL := /bin/bash
all:
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir"; \
	done

clean:
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir" clean; \
	done
