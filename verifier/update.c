// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <stdbool.h>
#include <string.h>
#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

#include "update.h"

#define MIN(a, b) ((a) <= (b) ? (a) : (b))
#define WRITE_SIZE 256
const uint32_t WRITE_ALIGN_MASK = (~(WRITE_SIZE - 1));

int update_init(struct update_ctx *ctx, size_t app_size, uint8_t app_digest[32],
		uint8_t app_signature[64])
{
	if (app_size == 0 || app_size > TK1_APP_MAX_SIZE) {
		debug_puts("Invalid app size!\n");
		return -1;
	}

	ctx->upload_offset = 0;
	ctx->upload_size = app_size;
	memcpy(ctx->app_digest, app_digest, sizeof(ctx->app_digest));
	memcpy(ctx->app_signature, app_signature, sizeof(ctx->app_signature));

	return sys_preload_delete();
}

int update_write(struct update_ctx *ctx, uint8_t *data, size_t data_len)
{
	assert(ctx->upload_size > ctx->upload_offset);
	data_len = MIN(ctx->upload_size - ctx->upload_offset, data_len);

	if (write_app(ctx->upload_offset, data, data_len) != 0) {
		return -1;
	}

	ctx->upload_offset += data_len;

	return 0;
}

bool update_app_is_written(struct update_ctx *ctx)
{
	return ctx->upload_offset >= ctx->upload_size;
}

int update_finalize(struct update_ctx *ctx)
{
	if (sys_preload_store_fin(ctx->upload_size, ctx->app_digest,
				  ctx->app_signature) != 0) {
		return -1;
	}

	ctx->upload_size = 0;
	memset(ctx->app_digest, 0, sizeof(ctx->app_digest));
	memset(ctx->app_signature, 0, sizeof(ctx->app_signature));

	return 0;
}

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
