// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include <tkey/syscall.h>

#include "fakesys.h"

#define APP_SLOT_1_SIZE (128 * 1024)

struct digsig {
	uint8_t digest[32];
	uint8_t signature[64];
};

struct digsig digsig = {0};
uint8_t app_slot_1[APP_SLOT_1_SIZE] = {0};

int sys_get_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(digest, digsig.digest, sizeof(digsig.digest));
	memcpy(signature, digsig.signature, sizeof(digsig.signature));

	return 0;
}

void fakesys_set_digsig(uint8_t digest[32], uint8_t signature[64])
{
	memcpy(digsig.digest, digest, sizeof(digsig.digest));
	memcpy(digsig.signature, signature, sizeof(digsig.signature));
}

int sys_preload_store(uint32_t offset, void *app, size_t len)
{
	if (offset >= sizeof(app_slot_1)) {
		return -1;
	}

	if (app == NULL) {
		return -1;
	}

	if (len > sizeof(app_slot_1)) {
		return -1;
	}

	_Static_assert(sizeof(app_slot_1) < (UINT32_MAX / 2), "avoid overflow");
	if ((offset + len) > sizeof(app_slot_1)) {
		return -1;
	}

	for (size_t i = 0; i < len; i++) {
		app_slot_1[offset + i] &= ((uint8_t *)app)[i];
	}

	return 0;
}

void fakesys_preload_erase(void)
{
	memset(app_slot_1, 0xff, sizeof(app_slot_1));
}

bool fakesys_preload_range_contains_ff(uint32_t start, uint32_t stop)
{
	if (start == stop) {
		return true;
	}

	if (start > stop) {
		return false;
	}

	if ((start < 0) || (start >= sizeof(app_slot_1))) {
		return false;
	}

	if ((stop < 0) || (stop > sizeof(app_slot_1))) {
		return false;
	}

	for (size_t i = start; i < stop; i++) {
		if (app_slot_1[i] != 0xff) {
			return false;
		}
	}

	return true;
}

bool fakesys_preload_range_contains_data(uint32_t offset, void *data,
					 size_t data_len)
{
	if (data_len > sizeof(app_slot_1)) {
		return false;
	}

	if (offset > sizeof(app_slot_1)) {
		return false;
	}

	_Static_assert(sizeof(app_slot_1) < (UINT32_MAX / 2), "avoid overflow");
	if ((offset + data_len) > sizeof(app_slot_1)) {
		return false;
	}

	for (size_t i = offset; i < data_len; i++) {
		uint8_t expected = *(((uint8_t *)data) + i);

		if (app_slot_1[i] != expected) {
			return false;
		}
	}

	return true;
}
