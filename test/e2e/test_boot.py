# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from .assets import testapp_a
from .commands import testapp_probe
from .tt.drivers.tkey import TKey


def test_should_boot_valid_app_in_flash_slot1(tkey: TKey) -> None:
    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
