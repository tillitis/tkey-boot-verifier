# SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
# SPDX-License-Identifier: BSD-2-Clause

import subprocess
from os import PathLike
from typing import Sequence

from ..tt.drivers.tkey import TKey


class TestappProbe:
    __test__ = False

    def __init__(
        self,
        cli_path: str | PathLike[str],
        tkey: TKey | None = None,
        extra_args: Sequence[str] = tuple(),
    ):
        self.cli_path = cli_path
        self.tkey = tkey
        self.extra_args = extra_args

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

        args.extend(self.extra_args)
        args.extend(extra_args)

        return subprocess.run(
            args, capture_output=True, check=True, text=True, timeout=timeout
        )

    def get_cdi(
        self,
        tkey: TKey | None,
        extra_args: Sequence[str] = tuple(),
        timeout: int | None = None,
    ) -> str:
        args = []
        args.extend(["-cmd", "get-cdi"])
        args.extend(extra_args)

        p = self.run(tkey, args, timeout=timeout)

        return p.stdout.strip()

    def get_nameversion(
        self,
        tkey: TKey | None,
        extra_args: Sequence[str] = tuple(),
        timeout: int | None = None,
    ) -> str:
        args = []
        args.extend(["-cmd", "get-nameversion"])
        args.extend(extra_args)

        p = self.run(tkey, args, timeout=timeout)

        return p.stdout.strip()
