# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

from subprocess import CalledProcessError

import pytest

from .assets import testapp_a, testapp_b
from .commands import testapp_probe, tkey_mgt
from .tt.drivers.tkey import TKey


def test_should_boot_valid_app_in_flash_slot1(tkey: TKey) -> None:
    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion


def test_can_install_another_app_in_slot_1(tkey: TKey) -> None:
    tkey_mgt.install_app(tkey, testapp_b.path, testapp_b.sig0_path)

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion

    tkey.eject()
    tkey.insert()

    assert testapp_probe.get_nameversion(tkey) == testapp_b.nameversion


def test_should_not_install_app_with_invalid_signature(tkey: TKey) -> None:
    with pytest.raises(CalledProcessError) as exc:
        tkey_mgt.install_app(tkey, testapp_b.path, testapp_b.sig1_path)

    assert exc.value.returncode != 0
    assert "couldn't update app slot 1: app signature invalid" in exc.value.stdout

    tkey.eject()
    tkey.insert()

    assert testapp_probe.get_nameversion(tkey) == testapp_a.nameversion
