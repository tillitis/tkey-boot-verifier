// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <monocypher/monocypher-ed25519.h>
#include <stdbool.h>
#include <tkey/assert.h>
#include <tkey/led.h>
#include <tkey/lib.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

// clang-format off
static volatile uint32_t *app_addr      = (volatile uint32_t *) TK1_MMIO_TK1_APP_ADDR;
static volatile uint32_t *app_size      = (volatile uint32_t *) TK1_MMIO_TK1_APP_SIZE;
static volatile uint32_t *cpu_mon_ctrl  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_CTRL;
static volatile uint32_t *cpu_mon_first = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_FIRST;
static volatile uint32_t *cpu_mon_last  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_LAST;
// clang-format on

int reset_if_verified(uint8_t pubkey[32])
{
	uint8_t app_digest[32];
	uint8_t app_signature[64];

	if (sys_get_digsig(app_digest, app_signature) != 0) {
		return -1;
	}

	if (crypto_ed25519_check(app_signature, pubkey, app_digest,
				 sizeof(app_digest)) != 0) {
		return -1;
	}

	// Reset to app slot 1 forcing check of app_digest
	struct reset rst = {0};

	rst.type = START_FLASH1_VER;
	memcpy_s(rst.app_digest, sizeof(rst.app_digest), app_digest,
		 sizeof(app_digest));

	sys_reset(&rst, 0); // will reset hardware!

	return -2;
}

int main(void)
{
	// Pubkey we got from tkeyimage
	// 9b62773323ef41a11834824194e55164d325eb9cdcc10ddda7d10ade4fbd8f6d
	uint8_t pubkey[32] = {
	    0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82,
	    0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1,
	    0x0d, 0xdd, 0xa7, 0xd1, 0x0a, 0xde, 0x4f, 0xbd, 0x8f, 0x6d,
	};

	// Use Execution Monitor on RAM after app
	*cpu_mon_first = *app_addr + *app_size;
	*cpu_mon_last = TK1_RAM_BASE + TK1_RAM_SIZE;
	*cpu_mon_ctrl = 1;

	reset_if_verified(pubkey);

	assert(1 == 2);
}
