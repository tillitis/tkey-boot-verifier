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

	buf[0] = genhdr(hdr.id, hdr.endpoint, 0x1, LEN_1);
	buf[1] = 0; // Not used, but smallest payload is 1 byte

	write(IO_CDC, buf, 2);
}

// Send app reply with frame header, response code, and LEN_X-1 bytes from buf
void appreply(struct frame_header hdr, enum appcmd rspcode, void *buf)
{
	size_t nbytes = 0; // Number of bytes in a reply frame
			   // (including rspcode).
	enum cmdlen len = LEN_1;
	uint8_t frame[1 + 128]; // Frame header + longest response

	switch (rspcode) {
	case CMD_GET_CDI:
		len = LEN_128;
		nbytes = 128;
		break;

	case CMD_GET_NAMEVERSION:
		len = LEN_32;
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
	frame[0] = genhdr(hdr.id, hdr.endpoint, 0x0, len);
	// App protocol header
	frame[1] = rspcode;

	// Copy payload after app protocol header
	memcpy(&frame[2], buf, nbytes - 1);

	write(IO_CDC, frame, 1 + nbytes);
}
