// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef APP_PROTO_H
#define APP_PROTO_H

#include <tkey/proto.h>

enum appcmd {
	CMD_GET_CDI = 0x01,
	CMD_RESET = 0xfe,
	CMD_FW_PROBE = 0xff,
};

void appreply_nok(struct frame_header hdr);
void appreply(struct frame_header hdr, enum appcmd rspcode, void *buf);

#endif
