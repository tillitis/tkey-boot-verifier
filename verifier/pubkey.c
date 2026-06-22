// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <blake2s/blake2s.h>
#include <monocypher/monocypher-ed25519.h>
#include <stdbool.h>
#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/syscall.h>

#include "app_proto.h"
#include "util.h"

void store_pubkey(struct packet pkt)
{
	uint8_t rsp[CMDLEN_MAXBYTES] = {0}; // Response

	if (pkt.hdr.len != 128) {
		// Bad length
		assert(1 == 2);
	}

	if (!user_is_present(1)) {
		rsp[0] = STATUS_BAD;
		appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
		return;
	}

	// Get the existing pubkey
	uint8_t app_digest[32] = {0};
	uint8_t app_signature[64] = {0};
	uint8_t pubkey[32] = {0};

	if (sys_preload_get_metadata(app_digest, app_signature, pubkey) != 0) {
		debug_puts("verifier:"
			   " sys_preload_get_metadata failed\n");
		rsp[0] = STATUS_BAD;
		appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
		return;
	}

	// The signature is over the digest of the new pubkey, so
	// compute the digest
	uint8_t digest[32] = {0};

	if (blake2s(digest, 32, NULL, 0, &pkt.cmd[1], 32) != 0) {
		debug_puts("verifier: couldn't do blake2s\n");
		rsp[0] = STATUS_BAD;
		appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
		return;
	}

	debug_puts("new pubkey:\n");
	debug_hexdump(&pkt.cmd[1], 32);

	debug_puts("signature:\n");
	debug_hexdump(&pkt.cmd[33], 64);

	debug_puts("digest over new pubkey:\n");
	debug_hexdump(digest, 32);

	// Verify the signature we just got pkt.cmd[33] over the
	// digest of new pubkey (pkt.cmd[1]) against already stored
	// pubkey
	if (crypto_ed25519_check(&pkt.cmd[33], pubkey, digest, 32) != 0) {
		debug_puts("verifier: signature verification failed\n");
		rsp[0] = STATUS_BAD;
		appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
		return;
	}

	if (sys_preload_set_pubkey(&pkt.cmd[1]) != 0) {
		rsp[0] = STATUS_BAD;
		appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
		return;
	}

	rsp[0] = STATUS_OK;
	appreply(pkt.hdr, CMD_STORE_PUBKEY, rsp);
}
