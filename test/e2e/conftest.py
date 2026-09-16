# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import shutil
from pathlib import Path
from typing import Iterator, List

import pytest

from .assets import pubkey0_path, testapp_a, verifier_path
from .device import new_qemu_tkey
from .drivers.tillitis_key1_repo import TK1Repo
from .tt.drivers.tkey import TKey, TKeyType


def pytest_addoption(parser: pytest.Parser) -> None:
    parser.addoption(
        "--qemu-path", default="qemu-system-riscv32", help="Path to QEmu executable"
    )

    parser.addoption(
        "--qemu-usb-mux-path",
        default="qemu_usb_mux.py",
        help="Path to TKey USB controller simulator",
    )

    parser.addoption(
        "--tk1-repo-path",
        required=True,
        type=Path,
        help="Path to tillitis-key1 repo. Will be used to generate assets",
    )


def pytest_report_header(config: pytest.Config) -> List[str]:
    qemu_path = Path(config.getoption("--qemu-path")).expanduser()
    qemu_usb_mux_path = Path(config.getoption("--qemu-usb-mux-path")).expanduser()
    tk1_repo_path = Path(config.getoption("--tk1-repo-path")).expanduser()

    return [
        f"qemu-path: {qemu_path}",
        f"qemu-usb-mux-path: {qemu_usb_mux_path}",
        f"tk1-repo-path: {tk1_repo_path}",
    ]


@pytest.fixture(scope="session")
def qemu_path(request: pytest.FixtureRequest) -> Path:
    return Path(request.config.getoption("--qemu-path")).expanduser()


@pytest.fixture(scope="session")
def qemu_usb_mux_path(request: pytest.FixtureRequest) -> Path:
    return Path(request.config.getoption("--qemu-usb-mux-path")).expanduser()


@pytest.fixture(scope="session")
def tk1_build(request: pytest.FixtureRequest) -> TK1Repo:
    path = Path(request.config.getoption("--tk1-repo-path")).expanduser()
    tk1_repo = TK1Repo(path)

    tk1_repo.build_qemu_firmware(verifier_path)
    tk1_repo.build_flash_image(
        verifier_path,
        testapp_a.path,
        pubkey0_path,
        testapp_a.sig0_path,
    )

    return tk1_repo


@pytest.fixture
def tkey(
    qemu_path: Path,
    qemu_usb_mux_path: Path,
    tk1_build: TK1Repo,
    tmp_path: Path,
) -> Iterator[TKey]:
    firmware_path = tmp_path / "qemu_firmware.elf"
    shutil.copy2(tk1_build.qemu_firmware_path, firmware_path)

    flash_image_path = tmp_path / "flash_image.bin"
    shutil.copy2(tk1_build.flash_image_path, flash_image_path)

    _tkey = new_qemu_tkey(
        TKeyType.CastorPre,
        qemu_path,
        qemu_usb_mux_path,
        firmware_path,
        flash_image_path,
    )
    _tkey.insert()
    yield _tkey
    _tkey.eject()
