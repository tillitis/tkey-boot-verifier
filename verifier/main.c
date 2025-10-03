// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <monocypher/monocypher-ed25519.h>
#include <stdbool.h>
#include <tkey/assert.h>
#include <tkey/led.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>
#include <tkey/proto.h>
#include <tkey/debug.h>
#include <tkey/platform.h>

#include "verify.h"
#include "app_proto.h"

// clang-format off
static volatile uint32_t *app_addr      = (volatile uint32_t *) TK1_MMIO_TK1_APP_ADDR;
static volatile uint32_t *app_size      = (volatile uint32_t *) TK1_MMIO_TK1_APP_SIZE;
static volatile uint32_t *cpu_mon_ctrl  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_CTRL;
static volatile uint32_t *cpu_mon_first = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_FIRST;
static volatile uint32_t *cpu_mon_last  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_LAST;
static volatile uint32_t *ver		= (volatile uint32_t *) TK1_MMIO_TK1_VERSION;
// clang-format on

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
			debug_puts("readselect errror");
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
		debug_puts("Couldn't parse header\n");
		return -1;
	}

	if (*ver >= TKEY_VERSION_CASTOR) {
		for (uint8_t n = 0; n < hdr->len;) {
			if (readselect(IO_CDC, &endpoint, &available) < 0) {
				debug_puts("readselect errror");
				return -1;
			}

			// Read as much as is available of what we expect from
			// the frame.
			available = available > hdr->len ? hdr->len : available;

			debug_puts("reading ");
			debug_putinthex(available);
			debug_lf();

			int nbytes = read(IO_CDC, &cmd[n], CMDLEN_MAXBYTES - n,
					  available);
			if (nbytes < 0) {
				debug_puts("read: buffer overrun\n");

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
		debug_puts("Responded NOK to message meant for fw\n");
		cmd[0] = CMD_FW_PROBE;

		return 0;
	}

	// Is it for us? If not, return error after having discarded
	// all bytes.
	if (hdr->endpoint != DST_SW) {
		debug_puts("Message not meant for app. endpoint was 0x");
		debug_puthex(hdr->endpoint);
		debug_lf();

		return -1;
	}

	return 0;
}

int main(void)
{
	uint8_t rsp[CMDLEN_MAXBYTES] = {0}; // Response
	size_t rsp_left =
	    CMDLEN_MAXBYTES; // How many bytes left in response buf
	struct packet pkt = {0};
	uint8_t app_digest[32];
	uint8_t app_signature[64];

	// Pubkey we got from tkeyimage
	// 9b62773323ef41a11834824194e55164d325eb9cdcc10ddda7d10ade4fbd8f6d
	uint8_t pubkey[32] = {
	    0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82,
	    0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1,
	    0x0d, 0xdd, 0xa7, 0xd1, 0x0a, 0xde, 0x4f, 0xbd, 0x8f, 0x6d,
	};

	// Use Execution Monitor on RAM after app
	*cpu_mon_first = *app_addr + *app_size;
	*cpu_mon_last = TK1_RAM_BASE + TK1_RAM_SIZE;
	*cpu_mon_ctrl = 1;

	// if started from client - wait for client data
	for (;;) {
		if (read_command(&pkt.hdr, pkt.cmd) != 0) {
			debug_puts("read_command returned != 0!\n");
			assert(1 == 2);
		}

		// Smallest possible payload length (cmd) is 1 byte.
		switch (pkt.cmd[0]) {
		case CMD_VERIFY:
			// read digest and sig from client
			memcpy(app_digest, &pkt.cmd[1], 32);
			memcpy(app_signature, &pkt.cmd[33], 64);
			reset_if_verified(pubkey, START_CLIENT_VER, app_digest, app_signature);
			assert(1 == 2);
			break;

		case CMD_RESET:
			assert(1 == 2);
			break;

		default:
			// WTF?
			assert(1 == 2);
			break;
		}
	}

	// if started from flash

	if (sys_get_digsig(app_digest, app_signature) != 0) {
		return -1;
	}

	reset_if_verified(pubkey, START_FLASH1_VER, app_digest, app_signature);

	assert(1 == 2);
}
