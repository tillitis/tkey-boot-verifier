// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <blake2s/blake2s.h>
#include <monocypher/monocypher-ed25519.h>
#include <tkey/assert.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>

#include "verify.h"

int reset_if_verified(uint8_t pubkey[32], enum reset_start reset_type,
		      uint8_t app_digest[32], uint8_t app_signature[64])
{
	if (crypto_ed25519_check(app_signature, pubkey, app_digest, 32) != 0) {
		return -1;
	}

	// Reset to app slot 1 forcing check of app_digest
	struct reset rst = {
		.type = reset_type,
		.mask = RESET_SEED,
		.seed_digest = {0},
		.app_digest = {0},
	};

	// Make a digest of our security policy and anything else we
	// want to measure to be part of the next app's identity.
	// Currently just the vendor public key.
	int rc = blake2s(rst.seed_digest, RESET_DIGEST_SIZE, NULL, 0, (const void *)pubkey, 32);
	assert(rc == 0);

	memcpy_s(rst.app_digest, sizeof(rst.app_digest), app_digest, 32);

	sys_reset(&rst, 0); // will reset hardware!

	return -2;
}
