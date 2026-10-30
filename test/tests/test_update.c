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
#include "../utils/tcmocka_util.h"

static void test_write_app_can_write_app_to_erased_slot(void **state);
static void test_write_app_should_only_write_app(void **state);
static void test_write_app_should_fail_when_app_is_too_large(void **state);
static void test_update_can_update_app(void **state);
static void test_update_should_fail_w_invalid_signature(void **state);
static void
test_update_should_not_erase_preloaded_on_invalid_input(void **state);

int main(void)
{
	const struct CMUnitTest tests[] = {
	    cmocka_unit_test(test_update_can_update_app),
	    cmocka_unit_test(test_update_should_fail_w_invalid_signature),
	    cmocka_unit_test(test_write_app_can_write_app_to_erased_slot),
	    cmocka_unit_test(test_write_app_should_only_write_app),
	    cmocka_unit_test(test_write_app_should_fail_when_app_is_too_large),
	    cmocka_unit_test(
		test_update_should_not_erase_preloaded_on_invalid_input),
	};

	return cmocka_run_group_tests(tests, NULL, NULL);
}

// Private key used to sign:
//   000102030405060708090a0b0c0d0e0f
//   000102030405060708090a0b0c0d0e0f
//   9b62773323ef41a11834824194e55164
//   d325eb9cdcc10ddda7d10ade4fbd8f6d
static const uint8_t SIGNATURE[64] = {
    0x47, 0xE5, 0xD7, 0xD8, 0x46, 0x36, 0x8E, 0x5C, 0x62, 0x35, 0xF2,
    0xCE, 0x73, 0x11, 0xCB, 0x8B, 0x8F, 0x8D, 0x61, 0xEF, 0x22, 0x87,
    0x60, 0x62, 0x53, 0xCE, 0xE8, 0xF5, 0x1D, 0xE8, 0x1B, 0x4E, 0x8E,
    0x81, 0x71, 0xFB, 0xC5, 0x07, 0xD7, 0xAD, 0x4C, 0x88, 0x3D, 0x77,
    0xB6, 0xBD, 0xED, 0x35, 0xC7, 0xCE, 0x5A, 0x7C, 0xA4, 0x0E, 0x89,
    0xD2, 0xEB, 0xC1, 0xED, 0x75, 0x06, 0xF6, 0x4C, 0x0E,
};
static const uint8_t INVALID_SIGNATURE[64] = {
    0x00, 0x9f, 0x49, 0x00, 0xf0, 0x93, 0xe9, 0xac, 0xed, 0x94, 0x64,
    0x62, 0x8e, 0xb7, 0x95, 0x45, 0x85, 0xb0, 0x27, 0x21, 0x5b, 0x0b,
    0x7f, 0xbf, 0x1d, 0xd7, 0x7f, 0x3b, 0xae, 0x43, 0x1e, 0x76, 0x01,
    0xad, 0xb1, 0xe5, 0x4d, 0xea, 0x85, 0x5a, 0xd6, 0xb2, 0xf8, 0x73,
    0x28, 0x38, 0xe6, 0xc4, 0x2f, 0x43, 0x94, 0x81, 0x4b, 0xd6, 0x6c,
    0xb4, 0x82, 0x85, 0x27, 0xf9, 0x2b, 0x2a, 0xbc, 0x0f,
};
static const uint8_t PUBKEY[32] = {
    0x9b, 0x62, 0x77, 0x33, 0x23, 0xef, 0x41, 0xa1, 0x18, 0x34, 0x82,
    0x41, 0x94, 0xe5, 0x51, 0x64, 0xd3, 0x25, 0xeb, 0x9c, 0xdc, 0xc1,
    0x0d, 0xdd, 0xa7, 0xd1, 0x0a, 0xde, 0x4f, 0xbd, 0x8f, 0x6d,
};
// App digest to generate_app(dst, dst_size, MAX_APP_SIZE)
static const uint8_t APP_DIGEST[32] = {
    0xB6, 0x97, 0x13, 0x6A, 0x9A, 0x53, 0x87, 0x8C, 0x2B, 0xD4, 0x6E,
    0xE4, 0x0C, 0x85, 0xC8, 0xB4, 0xCB, 0x4F, 0x81, 0x2E, 0xC7, 0xDC,
    0x3F, 0xE5, 0x46, 0x95, 0x8A, 0xC4, 0x6B, 0x5F, 0xA9, 0x06,
};

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

static void test_update_can_update_app(void **state)
{
	const size_t APP_SIZE = MAX_APP_SIZE;
	const size_t WRITE_SIZE = 127;
	uint8_t app_buf[MAX_APP_SIZE + 127] = {0};
	_Static_assert(sizeof(app_buf) >= APP_SIZE + WRITE_SIZE,
		       "app_buf must hold enough data for all writes");

	// Arrange
	fakesys_preload_erase();
	generate_app(app_buf, sizeof(app_buf), APP_SIZE);

	struct app_metadata metadata = {0};
	metadata.size = APP_SIZE;
	memcpy(metadata.digest, APP_DIGEST, sizeof(APP_DIGEST));
	memcpy(metadata.signature, SIGNATURE, sizeof(SIGNATURE));
	memcpy(metadata.pubkey, PUBKEY, sizeof(PUBKEY));
	fakesys_preload_set_metadata(&metadata);

	struct update_ctx ctx = {0};

	// Act and assert...
	int ret = 0;
	ret = update_init(&ctx, APP_SIZE, metadata.digest, metadata.signature);
	assert_int_equal(ret, 0);
	assert_false(update_app_is_written(&ctx));

	size_t offset = 0;
	while (offset < APP_SIZE - WRITE_SIZE) {
		ret = update_write(&ctx, &app_buf[offset], WRITE_SIZE);
		assert_int_equal(ret, 0);
		assert_false(update_app_is_written(&ctx));
		offset += WRITE_SIZE;
	}

	ret = update_write(&ctx, &app_buf[offset], WRITE_SIZE);
	assert_int_equal(ret, 0);
	assert_true(update_app_is_written(&ctx));

	ret = update_finalize(&ctx);
	assert_int_equal(ret, 0);

	assert_true(fakesys_preload_range_contains_data(0, app_buf, APP_SIZE));
}

static void test_update_should_fail_w_invalid_signature(void **state)
{
	const size_t APP_SIZE = MAX_APP_SIZE;
	const size_t WRITE_SIZE = 127;
	uint8_t app_buf[MAX_APP_SIZE + 127] = {0};
	_Static_assert(sizeof(app_buf) >= APP_SIZE + WRITE_SIZE,
		       "app_buf must hold enough data for all writes");

	// Arrange
	fakesys_preload_erase();
	generate_app(app_buf, sizeof(app_buf), APP_SIZE);

	struct app_metadata metadata = {0};
	metadata.size = APP_SIZE;
	memcpy(metadata.digest, APP_DIGEST, sizeof(APP_DIGEST));
	memcpy(metadata.signature, INVALID_SIGNATURE,
	       sizeof(INVALID_SIGNATURE));
	memcpy(metadata.pubkey, PUBKEY, sizeof(PUBKEY));
	fakesys_preload_set_metadata(&metadata);

	struct update_ctx ctx = {0};

	// Act and assert...
	int ret = 0;
	ret = update_init(&ctx, APP_SIZE, metadata.digest, metadata.signature);
	assert_int_equal(ret, -1);
}

static void
test_update_should_not_erase_preloaded_on_invalid_input(void **state)
{
	const size_t APP_SIZE = MAX_APP_SIZE;
	const size_t TOO_LARGE_APP_SIZE = MAX_APP_SIZE + MAX_APP_SIZE;
	const size_t WRITE_SIZE = 127;
	uint8_t app_buf[MAX_APP_SIZE + 127] = {0};
	_Static_assert(sizeof(app_buf) >= APP_SIZE + WRITE_SIZE,
		       "app_buf must hold enough data for all writes");

	// Arrange
	fakesys_preload_erase();
	generate_app(app_buf, sizeof(app_buf), APP_SIZE);

	struct app_metadata metadata = {0};
	metadata.size = APP_SIZE;
	memcpy(metadata.digest, APP_DIGEST, sizeof(APP_DIGEST));
	fakesys_set_app(app_buf, APP_SIZE, &metadata);

	struct update_ctx ctx = {0};

	// Act and assert...
	int ret = 0;
	ret = update_init(&ctx, TOO_LARGE_APP_SIZE, metadata.digest,
			  metadata.signature);
	assert_true(fakesys_preload_range_contains_data(0, app_buf, APP_SIZE));
	assert_int_equal(ret, -1);
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

		print_subtest("%lu, offset: 0x%x, len: 0x%lx\n", test_n,
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

static void test_write_app_should_fail_when_app_is_too_large(void **state)
{
	const struct test_case test_cases[] = {
	    {.offset = 0, .len = MAX_APP_SIZE + 1},
	    {.offset = 0, .len = MAX_APP_SIZE + 2},
	    {.offset = 0, .len = 1.42 * MAX_APP_SIZE},
	};

	size_t test_n = 0;

	for (test_n = 0; test_n < ARRAY_LEN(test_cases); test_n++) {
		const struct test_case *test = &test_cases[test_n];
		uint8_t app[128 * 1024];

		print_subtest("%lu, offset: %u, len: %lu\n", test_n,
			      test->offset, test->len);

		// Arrange
		fakesys_preload_erase();
		generate_app(app, sizeof(app), test->len);

		// Act
		int ret = write_app(test->offset, app, test->len);

		// Assert
		assert_int_equal(ret, -1);
	}
}
