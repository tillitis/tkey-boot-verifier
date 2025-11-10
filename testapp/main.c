// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <tkey/assert.h>
#include <tkey/debug.h>
#include <tkey/led.h>
#include <tkey/syscall.h>
#include <tkey/tk1_mem.h>

// clang-format off
static volatile uint32_t *app_addr      = (volatile uint32_t *) TK1_MMIO_TK1_APP_ADDR;
static volatile uint32_t *app_size      = (volatile uint32_t *) TK1_MMIO_TK1_APP_SIZE;
static volatile uint32_t *cdi           = (volatile uint32_t *)TK1_MMIO_TK1_CDI_FIRST;
static volatile uint32_t *cpu_mon_ctrl  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_CTRL;
static volatile uint32_t *cpu_mon_first = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_FIRST;
static volatile uint32_t *cpu_mon_last  = (volatile uint32_t *) TK1_MMIO_TK1_CPU_MON_LAST;
// clang-format on

#define CDI_SIZE 32

extern const uint8_t app_led_color;
extern const uint8_t app_name0[4];
extern const uint8_t app_name1[4];
const uint32_t app_version = 0x00000000;

int wait_byte()
{
	uint8_t in = 0;
	uint8_t available = 0;
	enum ioend endpoint = IO_NONE;

	if (readselect(IO_CDC, &endpoint, &available) < 0) {
		debug_puts("readselect error");
		return -1;
	}

	if (read(IO_CDC, &in, 1, 1) < 0) {
		debug_puts("read error");
		return -1;
	}

	return 0;
}

int main(void)
{
	led_set(app_led_color);

	for (int i = 0; i < sizeof(app_name0); i++) {
		debug_putchar(app_name0[i]);
	}
	for (int i = 0; i < sizeof(app_name1); i++) {
		debug_putchar(app_name1[i]);
	}
	debug_lf();

	// Use Execution Monitor on RAM after app
	*cpu_mon_first = *app_addr + *app_size;
	*cpu_mon_last = TK1_RAM_BASE + TK1_RAM_SIZE;
	*cpu_mon_ctrl = 1;

	if (wait_byte() != 0) {
		led_set(LED_RED);
		while (1)
			;
	}

	write(IO_CDC, app_name0, sizeof(app_name0));
	write(IO_CDC, app_name1, sizeof(app_name1));
	puts(IO_CDC, "\r\n");

	puts(IO_CDC, "CDI:\r\n");
	hexdump(IO_CDC, (void *)cdi, CDI_SIZE);
	puts(IO_CDC, "\r\n");

	if (wait_byte() != 0) {
		led_set(LED_RED);
		while (1)
			;
	}

	struct reset rst = {0};
	rst.type = START_DEFAULT;
	sys_reset(&rst, 0);
}
