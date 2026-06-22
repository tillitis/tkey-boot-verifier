// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

#include <stdbool.h>
#include <stdint.h>
#include <tkey/led.h>
#include <tkey/timer.h>
#include <tkey/touch.h>

#include "util.h"

#define PRESENCE_TIMEOUT_S 20
#define PRESENCE_REPEAT_DELAY_S 1

bool user_is_present(int c)
{
	for (uint8_t i = 0; i < c; i++) {
		bool present = touch_wait(APP_LED_COLOR, PRESENCE_TIMEOUT_S);
		if (!present) {
			return false;
		}
		led_set(LED_BLACK);
		timer_wait(PRESENCE_REPEAT_DELAY_S);
	}

	return true;
}

void signal_issue(void)
{
	for (uint8_t i = 0; i < 3; i++) {
		led_set(i % 2 ? APP_LED_COLOR : LED_BLACK);
		timer_wait(1);
	}
}
