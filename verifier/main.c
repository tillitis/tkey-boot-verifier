// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <monocypher/monocypher-ed25519.h>
#include <stdbool.h>
#include <string.h>
#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/led.h>
#include <tkey/lib.h>
#include <tkey/proto.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

#include "app_proto.h"
#include "bv_nad.h"
#include "update.h"
#include "verify.h"

// clang-format off
static volatile uint32_t *app_addr      = (volatile uint32_t *) TK1_MMIO_TK1_APP_ADDR;
static volatile uint32_t *app_size      = (volatile uint32_t *) TK1_MMIO_TK1_APP_SIZE;
static volatile uint32_t *cpu_mon_ctrl  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_CTRL;
static volatile uint32_t *cpu_mon_first = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_FIRST;
static volatile uint32_t *cpu_mon_last  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_LAST;
static volatile uint32_t *ver		= (volatile uint32_t *) TK1_MMIO_TK1_VERSION;
// clang-format on

#define TKEY_VERSION_CASTOR 6

#define CHUNK_PAYLOAD_LEN (CMDLEN_MAXBYTES - 1)

// Incoming packet from client
struct packet {
	struct frame_header hdr;      // Framing Protocol header
	uint8_t cmd[CMDLEN_MAXBYTES]; // Application level protocol
};

// read_command takes a frame header and a command to fill in after
// parsing. It returns 0 on success.
static int read_command(struct frame_header *hdr, uint8_t *cmd)
{
	uint8_t in = 0;
	uint8_t available = 0;
	enum ioend endpoint = IO_NONE;

	memset(hdr, 0, sizeof(struct frame_header));
	memset(cmd, 0, CMDLEN_MAXBYTES);

	if (*ver >= TKEY_VERSION_CASTOR) {
		if (readselect(IO_CDC, &endpoint, &available) < 0) {
			debug_puts("verifier: readselect errror");
			return -1;
		}

		if (read(IO_CDC, &in, 1, 1) < 0) {
			return -1;
		}
	} else {
		if (uart_read(&in, 1, 1) < 0) {
			return -1;
		}
	}

	if (parseframe(in, hdr) == -1) {
		debug_puts("verifier: Couldn't parse header\n");
		return -1;
	}

	if (*ver >= TKEY_VERSION_CASTOR) {
		for (uint8_t n = 0; n < hdr->len;) {
			if (readselect(IO_CDC, &endpoint, &available) < 0) {
				debug_puts("verifier: readselect errror");
				return -1;
			}

			// Read as much as is available of what we expect from
			// the frame.
			available = available > hdr->len ? hdr->len : available;

			debug_puts("verifier: reading ");
			debug_putinthex(available);
			debug_lf();

			int nbytes = read(IO_CDC, &cmd[n], CMDLEN_MAXBYTES - n,
					  available);
			if (nbytes < 0) {
				debug_puts("verifier: read: buffer overrun\n");

				return -1;
			}

			n += nbytes;
		}
	} else {
		if (uart_read(cmd, CMDLEN_MAXBYTES, hdr->len) < 0) {
			return -1;
		}
	}

	// Well-behaved apps are supposed to check for a client
	// attempting to probe for firmware. In that case destination
	// is firmware and we just reply NOK, discarding all bytes
	// already read.
	if (hdr->endpoint == DST_FW) {
		appreply_nok(*hdr);
		debug_puts("verifier: Responded NOK to message meant for fw\n");
		cmd[0] = CMD_FW_PROBE;

		return 0;
	}

	// Is it for us? If not, return error after having discarded
	// all bytes.
	if (hdr->endpoint != DST_SW) {
		debug_puts(
		    "verifier: Message not meant for app. endpoint was 0x");
		debug_puthex(hdr->endpoint);
		debug_lf();

		return -1;
	}

	return 0;
}

enum state {
	STATE_STARTED = 0,
	STATE_VERIFY_FLASH,
	STATE_WAIT_FOR_COMMAND,
	STATE_WAIT_FOR_APP_CHUNK,
};

// Context for the loading of a message
struct context {
	struct update_ctx update_ctx;
};

static enum state started(void)
{
	enum state state = STATE_STARTED;
	uint8_t next_app_data[RESET_DATA_SIZE] = {0};

	if (sys_reset_data(next_app_data) != 0) {
		assert(1 == 2);
	}

	if (next_app_data[0] == BV_NAD_WAIT_FOR_COMMAND) {
		state = STATE_WAIT_FOR_COMMAND;
	} else {
		state = STATE_VERIFY_FLASH;
	}

	return state;
}

void reset(uint32_t type, enum bv_nad reset_dst)
{
	if (reset_dst >= BV_NAD_COUNT) {
		assert(1 == 2);
	}

	struct reset rst = {0};
	rst.type = type;
	rst.next_app_data[0] = reset_dst;

	sys_reset(&rst, 1);
}

static enum state verify_flash(uint8_t app_digest[32],
			       uint8_t app_signature[64], uint8_t pubkey[32])
{
	led_set(LED_BLUE);

	reset_if_verified(pubkey, START_FLASH1_VER, app_digest, app_signature);

	return STATE_WAIT_FOR_COMMAND;
}

static void wait_for_app_chunk(struct context *ctx)
{
	struct packet pkt = {0};
	uint8_t rsp[CMDLEN_MAXBYTES] = {0}; // Response

	assert(ctx != NULL);

	if (read_command(&pkt.hdr, pkt.cmd) != 0) {
		debug_puts("verifier: read_command returned != 0!\n");
		assert(1 == 2);
	}

	switch (pkt.cmd[0]) {
	case CMD_UPDATE_APP_CHUNK:
		// Bad length
		if (pkt.hdr.len != CMDLEN_MAXBYTES) {
			assert(1 == 2);
		}

		if (update_write(&ctx->update_ctx, &pkt.cmd[1],
				 CHUNK_PAYLOAD_LEN) != 0) {
			assert(1 == 2);
		}

		rsp[0] = STATUS_OK;
		appreply(pkt.hdr, CMD_UPDATE_APP_CHUNK, rsp);

		if (update_app_is_written(&ctx->update_ctx)) {
			if (update_finalize(&ctx->update_ctx) != 0) {
				assert(1 == 2);
			}

			struct reset rst = {0};
			rst.type = START_DEFAULT;
			rst.next_app_data[0] = BV_NAD_BOOT_APP_1;
			sys_reset(&rst, 1);
		}

		break;

	default:
		assert(1 == 2);
	}
}

enum state wait_for_command(enum state state, struct context *ctx,
			    uint8_t pubkey[32])
{
	struct packet pkt = {0};
	uint8_t rsp[CMDLEN_MAXBYTES] = {0}; // Response

	assert(ctx != NULL);

	led_set(LED_GREEN);

	if (read_command(&pkt.hdr, pkt.cmd) != 0) {
		debug_puts("read_command returned != 0!\n");
		assert(1 == 2);
	}

	// Smallest possible payload length (cmd) is 1 byte.
	switch (pkt.cmd[0]) {
	case CMD_GET_PUBKEY:
		if (pkt.hdr.len != 1) {
			// Bad length
			assert(1 == 2);
		}

		memcpy_s(rsp, sizeof(rsp), pubkey, 32);

		appreply(pkt.hdr, CMD_GET_PUBKEY, rsp);
		break;

	case CMD_VERIFY: {
		uint8_t app_digest[32] = {0};
		uint8_t app_signature[64] = {0};
		// read digest and sig from client
		memcpy(app_digest, &pkt.cmd[1], 32);
		memcpy(app_signature, &pkt.cmd[33], 64);
		reset_if_verified(pubkey, START_CLIENT_VER, app_digest,
				  app_signature);
		break;
	}

	case CMD_RESET:
		if (pkt.hdr.len != 4) {
			assert(1 == 2);
		}

		reset(pkt.cmd[1], pkt.cmd[2]);
		break;

	case CMD_UPDATE_APP_INIT: {
		uint32_t app_size = 0;

		// Bad length
		if (pkt.hdr.len != CMDLEN_MAXBYTES) {
			assert(1 == 2);
		}

		// size, digest, signature
		// cmd[1..4] contains the size.
		app_size = pkt.cmd[1] + (pkt.cmd[2] << 8) + (pkt.cmd[3] << 16) +
			   (pkt.cmd[4] << 24);
		uint8_t *app_digest = &pkt.cmd[5];
		uint8_t *app_signature = &pkt.cmd[37];

		if (update_init(&ctx->update_ctx, app_size, app_digest,
				app_signature) != 0) {
			assert(1 == 2);
		}

		rsp[0] = STATUS_OK;
		appreply(pkt.hdr, CMD_UPDATE_APP_INIT, rsp);

		state = STATE_WAIT_FOR_APP_CHUNK;
		break;
	}

	default:
		// WTF?
		assert(1 == 2);
		break;
	}

	return state;
}

int main(void)
{
	debug_puts("verifier");
	debug_lf();

	struct context ctx = {0};
	enum state state = STATE_STARTED;
#ifdef BOOT_INTO_WAIT_FOR_COMMAND
	state = STATE_WAIT_FOR_COMMAND;
#endif
	uint8_t app_digest[32] = {0};
	uint8_t app_signature[64] = {0};
	uint8_t pubkey[32] = {0};

	// Use Execution Monitor on RAM after app
	*cpu_mon_first = *app_addr + *app_size;
	*cpu_mon_last = TK1_RAM_BASE + TK1_RAM_SIZE;
	*cpu_mon_ctrl = 1;

	if (sys_preload_get_metadata(app_digest, app_signature, pubkey) != 0) {
		debug_puts("verifier: sys_preload_get_metadata failed\n");
		assert(1 == 2);
	}

#ifdef TKEY_DEBUG
	config_endpoints(IO_CDC | IO_DEBUG);
#endif

	for (;;) {
		switch (state) {
		case STATE_STARTED:
			state = started();
			break;

		case STATE_VERIFY_FLASH: {
			debug_puts("verifier: STATE_WAIT_VERIFY_FLASH\n");
			state = verify_flash(app_digest, app_signature, pubkey);
			break;
		}

		case STATE_WAIT_FOR_COMMAND:
			debug_puts("verifier: STATE_WAIT_FOR_COMMAND\n");
			state = wait_for_command(state, &ctx, pubkey);
			break;

		case STATE_WAIT_FOR_APP_CHUNK:
			wait_for_app_chunk(&ctx);
			break;
		}
	}
}
