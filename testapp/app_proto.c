// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <string.h>
#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/io.h>
#include <tkey/lib.h>

#include "app_proto.h"

// Send reply frame with response status Not OK (NOK==1), shortest length
void appreply_nok(struct frame_header hdr)
{
	uint8_t buf[2];

	if (frame_gen_hdr(hdr.id, hdr.f_domain, 0x1, 1, &buf[0]) != 0) {
		debug_puts("frame_gen_hdr failed\n");
		assert(1 == 2);
	}
	buf[1] = 0; // Not used, but smallest payload is 1 byte

	write(IO_CDC, buf, 2);
}

// Send app reply with frame header, response code, and LEN_X-1 bytes from buf
void appreply(struct frame_header hdr, enum appcmd rspcode, void *buf)
{
	size_t nbytes = 0;	// Number of bytes in a reply frame
				// (including rspcode).
	uint8_t frame[1 + 128]; // Frame header + longest response

	switch (rspcode) {
	case CMD_GET_CDI:
		nbytes = 128;
		break;

	case CMD_GET_NAMEVERSION:
		nbytes = 32;
		break;

	default:
		debug_puts("appreply(): Unknown response code: ");
		debug_puthex(rspcode);
		debug_puts("\n");
		assert(1 == 2);
		return;
	}

	// Frame Protocol Header
	if (frame_gen_hdr(hdr.id, hdr.f_domain, 0x0, nbytes, &frame[0]) != 0) {
		debug_puts("frame_gen_hdr failed\n");
		assert(1 == 2);
	}
	// App protocol header
	frame[1] = rspcode;

	// Copy payload after app protocol header
	memcpy(&frame[2], buf, nbytes - 1);

	write(IO_CDC, frame, 1 + nbytes);
}
