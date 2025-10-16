CC = clang

LIBDIR = ../../tkey-libs
APPDIR = ../verifier

CMOCKA_CONFIG = \
	-DHAVE_MALLOC_H \
	-DHAVE_INTTYPES_H \
	-DHAVE_SIGNAL_H \
	-DHAVE_STRINGS_H \

CFLAGS = \
	-static -std=gnu99 -O2 -g -Wall -Werror=implicit-function-declaration \
	-Wno-incompatible-library-redeclaration \
	-I lib/cmocka/include -I $(LIBDIR) -I $(LIBDIR)/include \
	$(CMOCKA_CONFIG) \

LDFLAGS += \
	   -Wl,--wrap=sys_reset

TEST_CFILES= \
	lib/cmocka/src/cmocka.c \
	platform/assert_halt.c \
	platform/fakesys.c \
	$(LIBDIR)/libcommon/lib.c \
	$(LIBDIR)/monocypher/monocypher.c \
	$(LIBDIR)/monocypher/monocypher-ed25519.c \

all: tests/test_update tests/test_verify

TEST_UPDATE_CFILES = tests/test_update.c $(TEST_CFILES) $(APPDIR)/update.c
tests/test_update: $(TEST_UPDATE_CFILES) Makefile
	$(CC) $(CFLAGS) $(TEST_UPDATE_CFILES) $(LDFLAGS) -o $@

TEST_VERIFY_CFILES = tests/test_verify.c $(TEST_CFILES) $(APPDIR)/verify.c
tests/test_verify: $(TEST_VERIFY_CFILES) Makefile
	$(CC) $(CFLAGS) $(TEST_VERIFY_CFILES) $(LDFLAGS) -o $@

FMTFILES=tests/*.[ch] platform/*.[ch] utils/*.[ch]
.PHONY: fmt
fmt:
	clang-format --dry-run --ferror-limit=0 $(FMTFILES)
	clang-format --verbose -i $(FMTFILES)
.PHONY: checkfmt
checkfmt:
	clang-format --dry-run --ferror-limit=0 --Werror $(FMTFILES)

.PHONY: clean
clean:
	rm -f tests/test_update tests/test_verify
