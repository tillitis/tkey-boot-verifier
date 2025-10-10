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
	-I lib/cmocka/include -I $(LIBDIR) -I $(LIBDIR)/include \
	$(CMOCKA_CONFIG) \

LDFLAGS += \
	   -Wl,--wrap=sys_reset

CFILES= \
	tests/test_verify.c \
	lib/cmocka/src/cmocka.c \
	platform/assert_halt.c \
	platform/fakesys.c \
	$(LIBDIR)/libcommon/lib.c \
	$(LIBDIR)/monocypher/monocypher.c \
	$(LIBDIR)/monocypher/monocypher-ed25519.c \
	$(APPDIR)/verify.c \

tests/test_verify: $(CFILES) Makefile
	$(CC) $(CFLAGS) $(CFILES) $(LDFLAGS) -o $@

