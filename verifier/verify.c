// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <blake2s/blake2s.h>
#include <monocypher/monocypher-ed25519.h>
#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>

#include "verify.h"

int reset_if_verified(const uint8_t pubkey[32], const uint8_t app_name[16],
		      enum reset_start reset_type, const uint8_t app_digest[32],
		      const uint8_t app_signature[64], bool installable)
{

	// Build the message:
	// BLAKE2s(app digest || name || installable)
	uint8_t msg[32] = {0};
	uint8_t flags[2] = {0};
	blake2s_ctx b2s_ctx;
	(void)blake2s_init(&b2s_ctx, 32, NULL, 0);
	blake2s_update(&b2s_ctx, app_digest, 32);
	blake2s_update(&b2s_ctx, app_name, 16);
	if (installable) {
		flags[0] = 1;
	}
	blake2s_update(&b2s_ctx, flags, sizeof(flags));
	blake2s_final(&b2s_ctx, msg);

	if (crypto_ed25519_check(app_signature, pubkey, msg, sizeof(msg)) !=
	    0) {
		debug_puts("verifier: signature verification failed\n");
		return -1;
	}

	// Reset to app slot 1 forcing check of app_digest
	struct reset rst = {
	    .type = reset_type,
	    .mask = RESET_SEED,
	    .measured_id_seed = {0},
	    .app_digest = {0},
	};

	// Make a digest of our security policy to be part of the next app's
	// identity. Currently the BLAKE2s(public key || app_name)

	blake2s_ctx ctx;

	(void)blake2s_init(&ctx, 32, NULL, 0);
	blake2s_update(&ctx, pubkey, 32);
	blake2s_update(&ctx, app_name, 16);
	blake2s_final(&ctx, rst.measured_id_seed);

	memcpy_s(rst.app_digest, sizeof(rst.app_digest), app_digest, 32);

	sys_reset(&rst, 0); // will reset hardware!

	return -2;
}
