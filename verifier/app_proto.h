// SPDX-FileCopyrightText: 2022 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef APP_PROTO_H
#define APP_PROTO_H

#include <tkey/io.h>
#include <tkey/lib.h>
#include <tkey/proto.h>

enum appcmd {
	CMD_VERIFY = 0x01,
	CMD_UPDATE_APP_INIT = 0x03,
	CMD_UPDATE_APP_CHUNK = 0x04,
	CMD_GET_PUBKEY = 0x05,
	CMD_STORE_PUBKEY = 0x06,
	CMD_SET_PUBKEY = 0x07,
	CMD_ERASE_AREAS = 0x08,

	CMD_RESET = 0xfe,
	CMD_FW_PROBE = 0xff,
};

void appreply_nok(struct frame_header hdr);
void appreply(struct frame_header hdr, enum appcmd rspcode, void *buf);

#endif
