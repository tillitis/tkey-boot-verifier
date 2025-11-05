// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#ifndef APP_UPDATE_H
#define APP_UPDATE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

struct update_ctx {
	size_t upload_size;
	size_t upload_offset;
	uint8_t app_digest[32];
	uint8_t app_signature[64];
};

int write_app(uint32_t addr, uint8_t *data, size_t sz);
int update_init(struct update_ctx *ctx, size_t app_size, uint8_t app_digest[32],
		uint8_t app_signature[64]);
bool update_app_is_written(struct update_ctx *ctx);
int update_write(struct update_ctx *ctx, uint8_t *data, size_t data_len);
int update_finalize(struct update_ctx *ctx);

#endif
