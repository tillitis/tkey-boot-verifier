// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/led.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

#include "../verifier/bv_nad.h"
#include "app_proto.h"

// clang-format off
static volatile uint32_t *app_addr      = (volatile uint32_t *) TK1_MMIO_TK1_APP_ADDR;
static volatile uint32_t *app_size      = (volatile uint32_t *) TK1_MMIO_TK1_APP_SIZE;
static volatile uint32_t *cdi           = (volatile uint32_t *) TK1_MMIO_TK1_CDI_FIRST;
static volatile uint32_t *cpu_mon_ctrl  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_CTRL;
static volatile uint32_t *cpu_mon_first = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_FIRST;
static volatile uint32_t *cpu_mon_last  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_LAST;
// clang-format on

#define CDI_SIZE 32

extern const uint8_t app_led_color;
extern const uint8_t app_name0[4];
extern const uint8_t app_name1[4];
const uint32_t app_version = 0x00000000;

enum state {
	STATE_STARTED,
	STATE_FAILED,
};

struct packet {
	struct frame_header hdr;      // Framing Protocol header
	uint8_t cmd[CMDLEN_MAXBYTES]; // Application level protocol
};

void debug_putname(void)
{
	for (int i = 0; i < sizeof(app_name0); i++) {
		debug_putchar(app_name0[i]);
	}
	for (int i = 0; i < sizeof(app_name1); i++) {
		debug_putchar(app_name1[i]);
	}
	debug_puts(": ");
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

static enum state started_commands(enum state state, struct packet pkt)
{
	uint8_t rsp[CMDLEN_MAXBYTES] = {0}; // Response

	// Smallest possible payload length (cmd) is 1 byte.
	switch (pkt.cmd[0]) {
	case CMD_FW_PROBE:
		// Firmware probe. Allowed in this protocol state.
		// State unchanged.
		break;

	case CMD_GET_CDI:
		debug_putname();
		debug_puts("CMD_GET_CDI\n");
		if (pkt.hdr.len != 1) {
			// Bad length
			state = STATE_FAILED;
			break;
		}

		memcpy_s(rsp, sizeof(rsp), (void *)cdi, CDI_SIZE);

		appreply(pkt.hdr, CMD_GET_CDI, rsp);

		break;

	case CMD_RESET:
		debug_putname();
		debug_puts("CMD_RESET\n");

		if (pkt.hdr.len != 4) {
			debug_putname();
			debug_puts("unexpected pkt.hdr.len: 0x");
			debug_puthex(pkt.hdr.len);
			debug_lf();
			state = STATE_FAILED;
			break;
		}

		reset(pkt.cmd[1], pkt.cmd[2]);
		debug_putname();
		debug_puts("expected reset");

		state = STATE_FAILED;
		break;

	default:
		debug_putname();
		debug_puts("Got unknown initial command: 0x");
		debug_puthex(pkt.cmd[0]);
		debug_lf();

		state = STATE_FAILED;
		break;
	}

	return state;
}

static int read_command(struct frame_header *hdr, uint8_t *cmd)
{
	uint8_t in = 0;
	uint8_t available = 0;
	enum ioend endpoint = IO_NONE;

	if (readselect(IO_CDC, &endpoint, &available) < 0) {
		debug_putname();
		debug_puts("readselect error");
		return -1;
	}

	if (read(IO_CDC, &in, 1, 1) < 0) {
		debug_putname();
		debug_puts("read error");
		return -1;
	}

	if (parseframe(in, hdr) == -1) {
		debug_putname();
		debug_puts("Couldn't parse header\n");
		return -1;
	}

	for (uint8_t n = 0; n < hdr->len;) {
		if (readselect(IO_CDC, &endpoint, &available) < 0) {
			debug_putname();
			debug_puts("readselect errror");
			return -1;
		}

		// Read as much as is available of what we expect from
		// the frame.
		available = available > hdr->len ? hdr->len : available;

		int nbytes =
		    read(IO_CDC, &cmd[n], CMDLEN_MAXBYTES - n, available);
		if (nbytes < 0) {
			debug_putname();
			debug_puts("read: buffer overrun\n");

			return -1;
		}

		n += nbytes;
	}

	// Well-behaved apps are supposed to check for a client
	// attempting to probe for firmware. In that case destination
	// is firmware and we just reply NOK, discarding all bytes
	// already read.
	if (hdr->endpoint == DST_FW) {
		appreply_nok(*hdr);
		debug_putname();
		debug_puts("Responded NOK to message meant for fw\n");
		cmd[0] = CMD_FW_PROBE;

		return 0;
	}

	// Is it for us? If not, return error after having discarded
	// all bytes.
	if (hdr->endpoint != DST_SW) {
		debug_putname();
		debug_puts("Message not meant for app. endpoint was 0x");
		debug_puthex(hdr->endpoint);
		debug_lf();

		return -1;
	}

	return 0;
}

int main(void)
{
	enum state state = STATE_STARTED;

	led_set(app_led_color);

	debug_putname();
	debug_lf();

	// Use Execution Monitor on RAM after app
	*cpu_mon_first = *app_addr + *app_size;
	*cpu_mon_last = TK1_RAM_BASE + TK1_RAM_SIZE;
	*cpu_mon_ctrl = 1;

	for (;;) {
		struct packet pkt = {0};

		if (read_command(&pkt.hdr, pkt.cmd) != 0) {
			debug_putname();
			debug_puts("read_command returned != 0!\n");
			state = STATE_FAILED;
		}

		switch (state) {
		case STATE_STARTED:
			debug_putname();
			debug_puts("STATE_STARTED");
			debug_lf();
			state = started_commands(state, pkt);
			break;

		case STATE_FAILED:
			debug_putname();
			debug_puts("STATE_FAILED");
			debug_lf();
			assert(1 == 2);
			break; // Not reached

		default:
			debug_putname();
			debug_puts("Unknown state: 0x");
			debug_putinthex(state);
			debug_lf();
			state = STATE_FAILED;
		}
	}
}
