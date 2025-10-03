// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <monocypher/monocypher-ed25519.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>

#include "verify.h"

int reset_if_verified(uint8_t pubkey[32])
{
	uint8_t app_digest[32];
	uint8_t app_signature[64];

	if (sys_get_digsig(app_digest, app_signature) != 0) {
		return -1;
	}

	if (crypto_ed25519_check(app_signature, pubkey, app_digest,
				 sizeof(app_digest)) != 0) {
		return -1;
	}

	// Reset to app slot 1 forcing check of app_digest
	struct reset rst = {0};

	rst.type = START_FLASH1_VER;
	memcpy_s(rst.app_digest, sizeof(rst.app_digest), app_digest,
		 sizeof(app_digest));

	sys_reset(&rst, 0); // will reset hardware!

	return -2;
}
