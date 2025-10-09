// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/platform.h>
#include <tkey/tk1_mem.h>

#include "app_proto.h"

// clang-format off
static volatile uint32_t *ver		= (volatile uint32_t *) TK1_MMIO_TK1_VERSION;
// clang-format on

// Send reply frame with response status Not OK (NOK==1), shortest length
void appreply_nok(struct frame_header hdr)
{
	uint8_t buf[2];
	enum ioend dst = IO_CDC;

	buf[0] = genhdr(hdr.id, hdr.endpoint, 0x1, LEN_1);
	buf[1] = 0; // Not used, but smallest payload is 1 byte

	if (*ver < TKEY_VERSION_CASTOR) {
		dst = IO_UART;
	}

	write(dst, buf, 2);
}

// Send app reply with frame header, response code, and LEN_X-1 bytes from buf
void appreply(struct frame_header hdr, enum appcmd rspcode, void *buf)
{
	size_t nbytes = 0; // Number of bytes in a reply frame
			   // (including rspcode).
	enum cmdlen len = LEN_1;
	uint8_t frame[1 + 128];	 // Frame header + longest response
	enum ioend dst = IO_CDC; // I/O destination.

	switch (rspcode) {
	case CMD_VERIFY:
		len = LEN_4;
		nbytes = 4;
		break;

	case CMD_UPDATE_APP_INIT:
		len = LEN_4;
		nbytes = 4;
		break;

	case CMD_UPDATE_APP_CHUNK:
		len = LEN_4;
		nbytes = 4;
		break;

	default:
		debug_puts("appreply(): Unknown response code: ");
		debug_puthex(rspcode);
		debug_puts("\n");
		assert(1 == 2);
		return;
	}

	// Frame Protocol Header
	frame[0] = genhdr(hdr.id, hdr.endpoint, 0x0, len);
	// App protocol header
	frame[1] = rspcode;

	// Copy payload after app protocol header
	memcpy(&frame[2], buf, nbytes - 1);

	if (*ver < TKEY_VERSION_CASTOR) {
		dst = IO_UART;
	}

	write(dst, frame, 1 + nbytes);
}
