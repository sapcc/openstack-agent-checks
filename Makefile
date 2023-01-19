SHELL := /bin/bash
all:
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir"; \
	done

clean:
	if [ -d "bin" ]; then rm -rf bin; fi; \
	for dir in cmd/*; do \
		[[ -f $$dir/Makefile ]] && echo Entering into "$$dir" && $(MAKE) -C "$$dir" clean; \
	done

bin: all
	if ! [ -d "bin" ]; then mkdir bin; fi; \
	for dir in cmd/*; do \
		mv $$dir/$$(basename $$dir) bin/; \
	done
