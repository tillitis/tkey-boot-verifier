// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <tkey/io.h>

#include "app_proto.h"

// Send reply frame with response status Not OK (NOK==1), shortest length
void appreply_nok(struct frame_header hdr)
{
	uint8_t buf[2];

	buf[0] = genhdr(hdr.id, hdr.endpoint, 0x1, LEN_1);
	buf[1] = 0; // Not used, but smallest payload is 1 byte

	write(IO_CDC, buf, 2);
}
