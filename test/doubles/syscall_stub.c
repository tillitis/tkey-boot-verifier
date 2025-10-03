// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <stdint.h>
#include <string.h>
#include <tkey/syscall.h>

#include "syscall_stub.h"

struct digsig {
	uint8_t digest[32];
	uint8_t signature[64];
};

struct digsig digsig = {0};

int sys_get_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(digest, digsig.digest, sizeof(digsig.digest));
	memcpy(signature, digsig.signature, sizeof(digsig.signature));

	return 0;
}

void sys_stub_set_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(digsig.digest, digest, sizeof(digsig.digest));
	memcpy(digsig.signature, signature, sizeof(digsig.signature));
}
