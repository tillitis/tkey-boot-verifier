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
	struct app_metadata metadata;
	uint8_t app[APP_SLOT_1_SIZE];
};

struct state state = {0};

int sys_preload_get_metadata(struct app_metadata *metadata)
{
	memcpy(&metadata->size, &state.metadata.size,
	       sizeof(state.metadata.size));
	memcpy(metadata->digest, state.metadata.digest,
	       sizeof(state.metadata.digest));
	memcpy(metadata->signature, state.metadata.signature,
	       sizeof(state.metadata.signature));
	memcpy(metadata->pubkey, state.metadata.pubkey,
	       sizeof(state.metadata.pubkey));

	return 0;
}

void fakesys_preload_set_metadata(struct app_metadata *metadata)

{
	memcpy(&state.metadata.size, &metadata->size,
	       sizeof(state.metadata.size));
	memcpy(state.metadata.digest, metadata->digest,
	       sizeof(state.metadata.digest));
	memcpy(state.metadata.signature, metadata->signature,
	       sizeof(state.metadata.signature));
	memcpy(state.metadata.pubkey, metadata->pubkey,
	       sizeof(state.metadata.pubkey));
}

int sys_preload_delete(struct app_metadata *metadata)
{

	metadata->size = 0;
	memset(metadata->digest, 0, sizeof(metadata->digest));
	memset(metadata->signature, 0, sizeof(metadata->signature));
	fakesys_preload_set_metadata(metadata);
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

int sys_preload_set_metadata(struct app_metadata *metadata)
{
	if (metadata->size == 0 || metadata->size > sizeof(state.app)) {
		return -1;
	}

	fakesys_preload_set_metadata(metadata);

	return 0;
}

void fakesys_preload_erase(void)
{
	memset(state.metadata.digest, 0, sizeof(state.metadata.digest));
	memset(state.metadata.signature, 0, sizeof(state.metadata.signature));
	state.metadata.size = 0;
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
