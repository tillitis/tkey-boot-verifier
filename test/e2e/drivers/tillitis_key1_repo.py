# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import subprocess
from os import PathLike
from pathlib import Path


class TK1Repo:
    def __init__(self, path: str | PathLike[str]) -> None:
        self.path = Path(path)

        self._build_dir = self.path / "hw" / "application_fpga"
        self.qemu_firmware_path = self._build_dir / "qemu_firmware.elf"
        self.flash_image_path = self._build_dir / "flash_image.bin"

    def build_flash_image(
        self,
        app0_path: str | PathLike[str],
        app1_path: str | PathLike[str],
        app1_pub_path: str | PathLike[str],
        app1_sig_path: str | PathLike[str],
    ) -> None:
        cwd = self._build_dir

        subprocess.run(
            [
                "make",
                "flash_image.bin",
                f"FLASH_APP_0={str(Path(app0_path).absolute())}",
                f"FLASH_APP_1={str(Path(app1_path).absolute())}",
                f"FLASH_APP_1_PUB={str(Path(app1_pub_path).absolute())}",
                f"FLASH_APP_1_SIG={str(Path(app1_sig_path).absolute())}",
            ],
            cwd=cwd,
            check=True,
        )

    def build_qemu_firmware(self, app0_path: str | PathLike[str]) -> None:
        cwd = self._build_dir

        subprocess.run(
            [
                "make",
                "qemu_firmware.elf",
                f"FLASH_APP_0={str(Path(app0_path).absolute())}",
            ],
            cwd=cwd,
            check=True,
        )
