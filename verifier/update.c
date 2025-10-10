// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <tkey/debug.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>

#include "update.h"

#define WRITE_SIZE 256
const uint32_t WRITE_ALIGN_MASK = (~(WRITE_SIZE - 1));

int write_app(uint32_t addr, uint8_t *data, size_t sz)
{
	uint8_t buf[WRITE_SIZE];

	uint32_t buf_offset = addr & ~WRITE_ALIGN_MASK;
	size_t len = 0;

	debug_puts("app write addr=");
	debug_putinthex(addr);
	debug_puts(" size=");
	debug_putinthex(sz);

	for (size_t i = 0; i < sz; i += len) {
		size_t bytes_left = sz - i;
		size_t dst_max_len = WRITE_SIZE - buf_offset;
		size_t src_max_len =
		    bytes_left < WRITE_SIZE ? bytes_left : WRITE_SIZE;
		len = src_max_len < dst_max_len ? src_max_len : dst_max_len;

		memset(buf, 0xff, sizeof(buf));
		memcpy(buf + buf_offset, data + i, len);

		int ret =
		    sys_preload_store(addr & WRITE_ALIGN_MASK, buf, WRITE_SIZE);
		if (ret != 0) {
			debug_puts("write app failed, addr=");
			debug_putinthex(addr);
			debug_puts(" ret=");
			debug_putinthex(ret);
			return -1;
		}

		buf_offset = 0;
		addr += len;
	}

	return 0;
}
