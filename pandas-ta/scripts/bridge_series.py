from __future__ import annotations

import math
from typing import Any

import pandas as pd


def normalize_series(values: Any) -> pd.Series:
    """把输入数组稳定转成浮点序列，非法值统一落为 NaN。"""
    if not isinstance(values, list):
        raise ValueError("series 必须是数组")
    out: list[float] = []
    for item in values:
        if item is None:
            out.append(float("nan"))
            continue
        value = float(item)
        if not math.isfinite(value):
            out.append(float("nan"))
            continue
        out.append(value)
    return pd.Series(out, dtype="float64")


def build_series_payload(series: dict[str, Any]) -> dict[str, pd.Series]:
    """统一构造 pandas-ta 输入序列。

    关键约束：
    1. 所有数值序列必须共享同一套索引，避免 pandas-ta 内部逐列比较时报标签不一致。
    2. 当输入长度不一致时，统一按尾部对齐截断到最短长度，保证最近一段 K 线仍可参与计算。
    """
    raw_series: dict[str, pd.Series] = {}
    min_length: int | None = None
    for key, value in series.items():
        if not isinstance(value, list):
            continue
        normalized = normalize_series(value)
        raw_series[key] = normalized
        current_length = len(normalized)
        if min_length is None or current_length < min_length:
            min_length = current_length

    if min_length is None:
        return {}

    if min_length <= 0:
        shared_index = pd.RangeIndex(start=0, stop=0)
    else:
        end = pd.Timestamp.now(tz="UTC").tz_localize(None)
        shared_index = pd.date_range(end=end, periods=min_length, freq="min")

    aligned: dict[str, pd.Series] = {}
    for key, values in raw_series.items():
        if len(values) > min_length:
            values = values.iloc[-min_length:].reset_index(drop=True)
        else:
            values = values.reset_index(drop=True)
        values.index = shared_index
        aligned[key] = values
    return aligned
