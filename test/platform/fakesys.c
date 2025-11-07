// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

// clang-format off
#include <stdarg.h>
#include <stddef.h>
#include <setjmp.h>
#include <stdint.h>
#include <cmocka.h> // cmocka need to be included last
// clang-format on

#include <string.h>
#include <tkey/syscall.h>

#include "fakesys.h"

#define APP_SLOT_1_SIZE (128 * 1024)

struct state {
	uint8_t digest[32];
	uint8_t signature[64];
	uint32_t app_size;
	uint8_t app[APP_SLOT_1_SIZE];
};

struct state state = {0};

int sys_get_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(digest, state.digest, sizeof(state.digest));
	memcpy(signature, state.signature, sizeof(state.signature));

	return 0;
}

void fakesys_set_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(state.digest, digest, sizeof(state.digest));
	memcpy(state.signature, signature, sizeof(state.signature));
}

int sys_preload_delete(void)
{
	uint8_t digest[32] = {0};
	uint8_t signature[64] = {0};

	state.app_size = 0;
	fakesys_set_digsig(digest, signature);
	memset(state.app, 0xff, sizeof(state.app));

	return 0;
}

int sys_preload_store(uint32_t offset, void *app, size_t len)
{
	if (offset >= sizeof(state.app)) {
		return -1;
	}

	if (app == NULL) {
		return -1;
	}

	if (len > sizeof(state.app)) {
		return -1;
	}

	_Static_assert(sizeof(state.app) < (UINT32_MAX / 2), "avoid overflow");
	if ((offset + len) > sizeof(state.app)) {
		return -1;
	}

	for (size_t i = 0; i < len; i++) {
		state.app[offset + i] &= ((uint8_t *)app)[i];
	}

	return 0;
}

int sys_preload_store_fin(size_t len, uint8_t digest[32], uint8_t signature[64])
{
	if (len == 0 || len > sizeof(state.app)) {
		return -1;
	}

	fakesys_set_digsig(digest, signature);
	state.app_size = len;

	return 0;
}

void fakesys_preload_erase(void)
{
	memset(state.digest, 0, sizeof(state.digest));
	memset(state.signature, 0, sizeof(state.signature));
	state.app_size = 0;
	memset(state.app, 0xff, sizeof(state.app));
}

bool fakesys_preload_range_contains_ff(uint32_t start, uint32_t stop)
{
	if (start == stop) {
		return true;
	}

	if (start > stop) {
		return false;
	}

	if ((start < 0) || (start >= sizeof(state.app))) {
		return false;
	}

	if ((stop < 0) || (stop > sizeof(state.app))) {
		return false;
	}

	for (size_t i = start; i < stop; i++) {
		if (state.app[i] != 0xff) {
			cm_print_error(
			    "app slot differs at offset: %lu, expected: "
			    "0xff, got: %u\n",
			    i, state.app[i]);
			return false;
		}
	}

	return true;
}

bool fakesys_preload_range_contains_data(uint32_t offset, void *data,
					 size_t data_len)
{
	if (data_len > sizeof(state.app)) {
		return false;
	}

	if (offset > sizeof(state.app)) {
		return false;
	}

	_Static_assert(sizeof(state.app) < (UINT32_MAX / 2), "avoid overflow");
	if ((offset + data_len) > sizeof(state.app)) {
		return false;
	}

	for (size_t i = offset; i < data_len; i++) {
		uint8_t expected = *(((uint8_t *)data) + i);

		if (state.app[i] != expected) {
			cm_print_error(
			    "app slot differs at offset: %lu, expected: "
			    "%u, got: %u\n",
			    i, expected, state.app[i]);
			return false;
		}
	}

	return true;
}
