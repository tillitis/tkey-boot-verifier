# Check for OS, if not macos assume linux
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	shasum = shasum -a 512
else
	shasum = sha512sum
endif

IMAGE=ghcr.io/tillitis/tkey-builder:5rc2

OBJCOPY ?= llvm-objcopy

P := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
LIBDIR ?= $(P)/../tkey-libs

CC = clang

INCLUDE = $(LIBDIR)/include

# If you want libcommon's debug_puts() et cetera to output something
# on the QEMU debug port, use -DQEMU_DEBUG, or -DTKEY_DEBUG if you
# want it on the TKey HID debug endpoint
CFLAGS = -target riscv32-unknown-none-elf -march=rv32iczmmul -mabi=ilp32 -mcmodel=medany \
   -static -std=gnu99 -O2 -ffast-math -fno-common -fno-builtin-printf \
   -fno-builtin-putchar -nostdlib -mno-relax -flto -g \
   -Wall -Werror=implicit-function-declaration \
   -I $(INCLUDE) -I $(LIBDIR) #-DTKEY_DEBUG #-DQEMU_DEBUG

AS = clang
ASFLAGS = -target riscv32-unknown-none-elf -march=rv32iczmmul -mabi=ilp32 -mcmodel=medany -mno-relax

LDFLAGS=-T $(LIBDIR)/app.lds -L $(LIBDIR) -lcommon -lcrt0


.PHONY: all
all: verifier/app.bin check-verifier-hash

# Create compile_commands.json for clangd and LSP
.PHONY: clangd
clangd: compile_commands.json
compile_commands.json:
	$(MAKE) clean
	bear -- make verifier/app.bin

# Turn elf into bin for device
%.bin: %.elf
	$(OBJCOPY) --input-target=elf32-littleriscv --output-target=binary $^ $@
	chmod a-x $@

show-%-hash: %/app.bin
	@echo "Device app digest:"
	@$(shasum) $$(dirname $^)/app.bin

check-verifier-hash: verifier/app.bin show-verifier-hash
	@echo "Expected device app digest: "
	@cat verifier/app.bin.sha512
	$(shasum) -c verifier/app.bin.sha512

.PHONY: check
check:
	clang-tidy -header-filter=.* -checks=cert-* verifier/*.[ch] -- $(CFLAGS)

# Simple ed25519 verifier app
VERIFIEROBJS=verifier/main.o
verifier/app.elf: $(VERIFIEROBJS)
	$(CC) $(CFLAGS) $(VERIFIEROBJS) $(LDFLAGS) -I $(LIBDIR) -o $@
$(VERIFIEROBJS): $(INCLUDE)/tkey/tk1_mem.h

.PHONY: clean
clean:
	rm -f verifier/app.bin verifier/app.elf $(VERIFIEROBJS)

# Uses ../.clang-format
FMTFILES=verifier/*.[ch]

.PHONY: fmt
fmt:
	clang-format --dry-run --ferror-limit=0 $(FMTFILES)
	clang-format --verbose -i $(FMTFILES)
.PHONY: checkfmt
checkfmt:
	clang-format --dry-run --ferror-limit=0 --Werror $(FMTFILES)

.PHONY: podman
podman:
	podman run --arch=amd64 --rm --mount type=bind,source=$(CURDIR),target=/src --mount type=bind,source=$(LIBDIR),target=/tkey-libs -w /src -it $(IMAGE) make -j
