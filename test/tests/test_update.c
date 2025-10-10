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

#include "../../verifier/update.h"
#include "../platform/fakesys.h"

static void test_write_app_can_write_app_to_erased_slot(void **state);
static void test_write_app_should_only_write_app(void **state);

int main(void)
{
	const struct CMUnitTest tests[] = {
	    cmocka_unit_test(test_write_app_can_write_app_to_erased_slot),
	    cmocka_unit_test(test_write_app_should_only_write_app),
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}

#define MAX_APP_SIZE (128 * 1024)
#define ARRAY_LEN(array) (sizeof(array) / (sizeof(array[0])))

struct test_case {
	uint32_t offset;
	size_t len;
};

int generate_app(void *dst, size_t dst_size, size_t app_len)
{
	if (app_len > dst_size) {
		return -1;
	}

	memset(dst, 0xA0, app_len);

	return 0;
}

static void test_write_app_can_write_app_to_erased_slot(void **state)
{
	uint8_t app[128 * 1024];

	// Arrange
	fakesys_preload_erase();
	generate_app(app, sizeof(app), sizeof(app));

	// Act
	int ret = write_app(0, app, sizeof(app));

	// Assert
	assert_int_equal(ret, 0);
	assert_true(fakesys_preload_range_contains_data(0, app, sizeof(app)));
}

static void test_write_app_should_only_write_app(void **state)
{
	const struct test_case test_cases[] = {
	    {.offset = 0, .len = 0},
	    {.offset = 0, .len = 1},
	    {.offset = 0, .len = 127},
	    {.offset = 0, .len = 128},
	    {.offset = 0, .len = 255},
	    {.offset = 0, .len = 256},
	    {.offset = 0, .len = MAX_APP_SIZE},
	    {.offset = 0, .len = MAX_APP_SIZE - 1},
	    {.offset = 1, .len = 0},
	    {.offset = 1, .len = 1},
	    {.offset = 1, .len = 2},
	    {.offset = 1, .len = 126},
	    {.offset = 1, .len = 127},
	    {.offset = 1, .len = 128},
	    {.offset = 1, .len = 129},
	    {.offset = 1, .len = 254},
	    {.offset = 1, .len = 255},
	    {.offset = 1, .len = 256},
	    {.offset = 1, .len = 257},
	    {.offset = 1, .len = MAX_APP_SIZE - 1},
	    {.offset = MAX_APP_SIZE / 2, .len = MAX_APP_SIZE / 2},
	};

	size_t test_n = 0;

	for (test_n = 0; test_n < ARRAY_LEN(test_cases); test_n++) {
		const struct test_case *test = &test_cases[test_n];
		uint8_t app[128 * 1024];

		print_message("test: %lu, offset: 0x%x, len: 0x%lx\n", test_n,
			      test->offset, test->len);

		// Arrange
		fakesys_preload_erase();
		generate_app(app, sizeof(app), test->len);

		// Act
		int ret = write_app(test->offset, app, test->len);

		// Assert
		assert_int_equal(ret, 0);
		assert_true(fakesys_preload_range_contains_ff(0, test->offset));
		assert_true(fakesys_preload_range_contains_data(
		    test->offset, app, test->len));
		assert_true(fakesys_preload_range_contains_ff(
		    test->offset + test->len, sizeof(app)));
	}
}
