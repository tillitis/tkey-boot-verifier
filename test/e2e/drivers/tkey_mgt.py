# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import subprocess
from os import PathLike
from typing import Sequence

from ..tt.drivers.qemu_tkey import QEmuTKey
from ..tt.drivers.tkey import TKey


class TKeyMgt:
    def __init__(
        self,
        cli_path: str | PathLike[str],
        tkey: TKey | None = None,
        extra_args: Sequence[str] = tuple(),
        cwd: str | PathLike[str] | None = None,
    ):
        self.cli_path = cli_path
        self.tkey = tkey
        self.extra_args = extra_args
        self.cwd = cwd

    def run(
        self,
        tkey: TKey | None,
        extra_args: Sequence[str] = tuple(),
        timeout: int | None = None,
    ) -> subprocess.CompletedProcess[str]:
        args = [self.cli_path]

        if tkey is not None:
            args.extend(["--port", tkey.cdc_port()])
        elif self.tkey is not None:
            args.extend(["--port", self.tkey.cdc_port()])

        if isinstance(tkey, QEmuTKey):
            args.append("--no-expect-close")

        args.extend(self.extra_args)
        args.extend(extra_args)

        return subprocess.run(
            args, capture_output=True, check=True, text=True, timeout=timeout
        )

    def boot(
        self,
        tkey: TKey | None,
        app_path: str | PathLike[str],
        sig_path: str | PathLike[str],
        pub_path: str | PathLike[str],
        extra_args: Sequence[str] = tuple(),
        timeout: int | None = 10,
    ) -> None:
        args = []
        args.extend(["-cmd", "boot"])
        args.extend(["-app", str(app_path)])
        args.extend(["-sig", str(sig_path)])
        args.extend(["-pub", str(pub_path)])
        args.extend(extra_args)

        self.run(tkey, args, timeout)

    def install_app(
        self,
        tkey: TKey | None,
        app_path: str | PathLike[str],
        sig_path: str | PathLike[str],
        extra_args: Sequence[str] = tuple(),
        timeout: int = 10,
    ) -> None:
        args = []
        args.extend(["-cmd", "install"])
        args.extend(["-app", str(app_path)])
        args.extend(["-sig", str(sig_path)])
        args.extend(extra_args)

        self.run(tkey, args, timeout)

    def install_pubkey(
        self,
        tkey: TKey | None,
        pubkey_path: str | PathLike[str],
        extra_args: Sequence[str] = tuple(),
        timeout: int | None = None,
    ) -> None:
        args = []
        args.extend(["-cmd", "install-pubkey"])
        args.extend(["-pub", str(pubkey_path)])
        args.extend(extra_args)

        self.run(tkey, args, timeout)
