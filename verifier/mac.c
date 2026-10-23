// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include "tkey/lib.h"
#include <blake2s/blake2s.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

// clang-format off
static volatile uint32_t *cdi = (volatile uint32_t *) TK1_MMIO_TK1_CDI_FIRST;
// clang-format on

// Calculates a MAC over the metadata, returns the result in out_mac.
// Takes the key as an input.
void mac_calc_key(const uint8_t key[32], const struct app_metadata *metadata,
		  uint8_t out_mac[32])
{
	blake2s_ctx ctx;

	(void)blake2s_init(&ctx, 32, key, 32);
	blake2s_update(&ctx, &metadata->size, sizeof(metadata->size));
	blake2s_update(&ctx, metadata->digest, sizeof(metadata->digest));
	blake2s_update(&ctx, metadata->signature, sizeof(metadata->signature));
	blake2s_update(&ctx, metadata->pubkey, sizeof(metadata->pubkey));
	blake2s_final(&ctx, out_mac);

	secure_wipe(&ctx, sizeof(ctx));
}

// Calculates a MAC over the metadata, returns the result in out_mac.
// Uses the CDI as the key.
void mac_calc(const struct app_metadata *metadata, uint8_t out_mac[32])
{
	mac_calc_key((const uint8_t *)cdi, metadata, out_mac);
}

// Returns true if MAC verifies.
bool mac_verify_key(const uint8_t key[32], const struct app_metadata *metadata)
{
	uint8_t mac[32];

	mac_calc_key(key, metadata, mac);
	bool valid = memeq(mac, metadata->mac, sizeof(mac)) != 0;

	return valid;
}

// Returns true if MAC verifies. Uses CDI as key.
bool mac_verify(const struct app_metadata *metadata)
{
	return mac_verify_key((const uint8_t *)cdi, metadata);
}
