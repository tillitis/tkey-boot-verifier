# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import subprocess
from os import PathLike
from typing import Optional, Sequence

from ..tt.drivers.tkey import TKey


class TKeyRunapp:
    def __init__(
        self,
        cli_path: str | PathLike[str],
        tkey: TKey | None = None,
        extra_args: Sequence[str] = tuple(),
        cwd: Optional[str | PathLike[str]] = None,
    ):
        self.cli_path = str(cli_path)
        self.tkey = tkey
        self.extra_args = tuple(extra_args)
        self.cwd = None if cwd is None else str(cwd)

    def run(
        self, tkey: TKey | None, extra_args: Sequence[str] = tuple()
    ) -> subprocess.CompletedProcess[bytes]:
        args = [self.cli_path]

        if tkey is not None:
            args.extend(["--port", tkey.cdc_port()])
            args.extend(["--speed", str(tkey.cdc_speed())])
        elif self.tkey is not None:
            args.extend(["--port", self.tkey.cdc_port()])
            args.extend(["--speed", str(self.tkey.cdc_speed())])

        args.extend(self.extra_args)
        args.extend(extra_args)

        return subprocess.run(args, check=True, cwd=self.cwd)

    def load(
        self,
        tkey: TKey | None,
        app_path: str | PathLike[str],
        extra_args: Sequence[str] = tuple(),
        uss_file_path: str | PathLike[str] | None = None,
        force_full_uss: bool = False,
    ) -> None:
        args = [str(app_path)]

        if uss_file_path is not None:
            args.extend(["--uss-file", str(uss_file_path)])
        if force_full_uss:
            args.extend(["--force-full-uss"])

        args.extend(extra_args)

        self.run(tkey, args)
