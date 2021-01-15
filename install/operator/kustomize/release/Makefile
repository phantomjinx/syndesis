
VERSION := 1.0.0

#
# Make Options
#
MK_OPTIONS := -s

.PHONY: $(SUBDIRS)

SYNCDIRS := ../app ../bases/crd ../bases/deployment ../bases/grant ../bases/jaeger-crd ../bases/role ../dev/app ../dev/operator

sync:
	for dir in $(SYNCDIRS); do \
    $(MAKE) $(MK_OPTIONS) -C $$dir; \
  done

RELEASEDIRS := ../app ../bases ../dev ../operator ../product ../setup ../setup-jaeger

syndesis:
	#@ Make a new syndesis builddirectory
	mkdir -p syndesis
	#@ Copy directories into build directory
	for dir in $(RELEASEDIRS); do \
    cp -rf $$dir syndesis/; \
  done
	#@ Remove sync rules from Makefiles
	for f in `find ./syndesis -name Makefile`; do \
		echo -e "sync:\n  #@ Not required in release" > $$f; \
	done
	#@ Copy the root Makefile into build directory
	cp -f ../Makefile syndesis/
	#@ Hand off to the release script
	./release.sh --version $(VERSION)

release: sync syndesis
ifndef VERSION
	$(error VERSION is not set)
endif

	tar zcvf syndesis-install-release-$(VERSION).tar.gz syndesis

clean:
	rm -rf syndesis *.tar.gz
