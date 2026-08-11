from __future__ import annotations

import json
import inspect
import math
import sys
from collections.abc import Iterable
from pathlib import Path
from typing import Any, get_args, get_origin

import pandas as pd
from bridge_series import build_series_payload


ROOT = Path(__file__).resolve().parents[3]


GROUP_NAME_MAPPING: dict[str, str] = {
    "cycles": "cycle_indicators",
    "statistics": "statistic_functions",
    "momentum": "momentum_indicators",
    "trend": "trend_indicators",
    "volatility": "volatility_indicators",
    "candles": "candle_patterns",
    "performance": "performance_metrics",
    "overlap": "overlap_studies",
    "volume": "volume_indicators",
}

GROUP_LABEL_MAPPING: dict[str, str] = {
    "cycles": "Cycle Indicators",
    "statistics": "Statistic Functions",
    "momentum": "Momentum Indicators",
    "trend": "Trend Indicators",
    "volatility": "Volatility Indicators",
    "candles": "Candle Patterns",
    "performance": "Performance Metrics",
    "overlap": "Overlap Studies",
    "volume": "Volume Indicators",
}

GROUP_PURPOSE_MAP: dict[str, str] = {
    "cycle_indicators": "分析市场周期长度、周期相位和趋势/周期切换",
    "trend_indicators": "判断趋势方向、强度和转折",
    "volatility_indicators": "衡量价格波动幅度和波动强度",
    "momentum_indicators": "衡量价格变化速度、强弱和动量转折",
    "overlap_studies": "对价格进行平滑、跟踪和通道分析，辅助观察趋势",
    "candle_patterns": "识别 K 线形态和蜡烛图反转信号",
    "volume_indicators": "结合成交量评估资金参与度和价格推动力",
    "performance_metrics": "衡量收益、回撤和表现质量",
    "statistic_functions": "做统计归纳、离散度和集中趋势分析",
}

MAIN_PANE_GROUPS = {"overlap", "trend", "volatility"}

SUMMARY_TRANSLATION_MAP: list[tuple[str, str]] = [
    ("Average True Range", "平均真实波幅"),
    ("Relative Strength Index", "相对强弱指标"),
    ("Moving average with variable period", "可变周期移动平均"),
    ("Moving average", "移动平均"),
    ("Exponential Moving Average", "指数移动平均"),
    ("Simple Moving Average", "简单移动平均"),
    ("Bollinger Bands", "布林带"),
    ("Trend", "趋势"),
    ("Volatility", "波动率"),
    ("Momentum", "动量"),
    ("Candle", "K 线"),
    ("Pattern", "形态"),
    ("Cycle", "周期"),
    ("Transform", "变换"),
]

FUNCTION_FORMULA_HINTS: dict[str, str] = {
    "ATR": "TR = max(high-low, |high-prev_close|, |low-prev_close|)；ATR = MA(TR, n)",
    "RSI": "RSI = 100 - 100 / (1 + RS)；RS = EMA(gain, n) / EMA(loss, n)",
    "MAVP": "MAVP(i) = MA(close, period[i])，按逐点变化的周期长度动态计算均线",
    "HT_DCPERIOD": "基于 Hilbert Transform 估计主导周期长度",
    "HT_DCPHASE": "基于 Hilbert Transform 估计主导周期相位",
    "HT_PHASOR": "基于 Hilbert Transform 的同相/正交分量解析",
    "HT_SINE": "基于主导相位生成正弦波与领先正弦波",
    "HT_TRENDMODE": "根据希尔伯特变换分解结果判断趋势/周期模式",
    "CDL2CROWS": "依据开高低收和前后蜡烛的组合规则识别两只乌鸦形态，输出 ±100 或 0",
    "CDL3BLACKCROWS": "依据连续三根阴线的组合规则识别三只黑乌鸦形态，输出 ±100 或 0",
    "CDL3INSIDE": "依据内包/外包关系识别三内形态，输出 ±100 或 0",
    "CDL3LINESTRIKE": "依据四根蜡烛的反转组合识别三线打击，输出 ±100 或 0",
    "CDL3OUTSIDE": "依据吞没后确认的组合识别三外形态，输出 ±100 或 0",
    "CDL3STARSINSOUTH": "依据三根低位止跌蜡烛的组合识别三星在南，输出 ±100 或 0",
    "CDL3WHITESOLDIERS": "依据连续三根阳线的组合识别三白兵，输出 ±100 或 0",
}

FUNCTION_DETAIL_OVERRIDES: dict[str, str] = {
    "ATR": "平均真实波幅，用于衡量价格波动强度，不区分方向，常用于止损和仓位控制。",
    "RSI": "相对强弱指标，用于观察价格上涨和下跌的相对强弱，常用于识别超买超卖。",
    "MAVP": "可变周期移动平均，按每个样本点指定的周期长度计算动态均线。",
    "HT_DCPERIOD": "希尔伯特变换主导周期长度，用于估计当前市场节奏是否在拉长或缩短。",
    "HT_DCPHASE": "希尔伯特变换主导周期相位，用于观察周期推进阶段和潜在拐点。",
    "HT_PHASOR": "希尔伯特变换相量分量，用于分析同相和正交分量，辅助识别周期转折。",
    "HT_SINE": "希尔伯特变换正弦波，用于观察周期节奏和领先拐点。",
    "HT_TRENDMODE": "希尔伯特变换趋势模式判断，用于区分当前更偏趋势还是周期震荡。",
    "HT_TRENDLINE": "希尔伯特变换趋势线，用于提取平滑的瞬时趋势主线。",
    "CDL2CROWS": "两只乌鸦形态，通常用于识别上涨后的短线转弱信号。",
    "CDL3BLACKCROWS": "三只黑乌鸦形态，通常用于识别连续走弱的看跌延续信号。",
    "CDL3INSIDE": "三内形态，用于观察收敛后可能出现的反转或延续。",
    "CDL3LINESTRIKE": "三线打击形态，用于识别强势反转后的情绪切换。",
    "CDL3OUTSIDE": "三外形态，用于识别吞没确认后的延续或反转。",
    "CDL3STARSINSOUTH": "三星在南形态，用于识别弱势末端的止跌观察结构。",
    "CDL3WHITESOLDIERS": "三白兵形态，用于识别连续上攻后的强势延续信号。",
}

FUNCTION_INPUT_REQUIREMENTS: dict[str, list[str]] = {
    "ATR": ["high", "low", "close"],
    "AO": ["high", "low"],
    "VWAP": ["high", "low", "close", "volume"],
    "MFI": ["high", "low", "close", "volume"],
    "OBV": ["close", "volume"],
    "CCI": ["high", "low", "close"],
    "WILLR": ["high", "low", "close"],
    "STOCH": ["high", "low", "close"],
    "STOCHRSI": ["close"],
    "MACD": ["close"],
    "BBANDS": ["close"],
    "ADX": ["high", "low", "close"],
    "ADXR": ["high", "low", "close"],
    "AROON": ["high", "low"],
    "CPR": ["open", "high", "low", "close"],
    "LONG_RUN": ["fast", "slow"],
    "SHORT_RUN": ["fast", "slow"],
    "TSIGNALS": ["trend"],
    "XSIGNALS": ["signal"],
    "CMO": ["close"],
    "MOM": ["close"],
    "RSI": ["close"],
    "MAVP": ["close", "periods"],
    "CDL_PATTERN": ["open", "high", "low", "close"],
    "CDL_Z": ["open", "high", "low", "close"],
    "HA": ["open", "high", "low", "close"],
}

FUNCTION_PARAMETER_DEFAULTS: dict[str, dict[str, Any]] = {
    "AO": {"fast": 5, "slow": 34},
    "VWAP": {"anchor": "D"},
    "TSIGNALS": {"asbool": False},
    "XSIGNALS": {"xa": 0, "xb": 1, "above": True, "long": True, "asbool": False},
    "LONG_RUN": {"length": 2},
    "SHORT_RUN": {"length": 2},
    "CPR": {"method": "classic", "timeframe": "daily", "levels": "standard"},
}

SIGNATURE_INPUT_ALIASES: dict[str, str] = {
    "open": "open",
    "open_": "open",
    "high": "high",
    "low": "low",
    "close": "close",
    "benchmark": "benchmark",
    "volume": "volume",
    "periods": "periods",
}


def _read_stdin_json() -> dict[str, Any]:
    raw = sys.stdin.read().strip()
    if not raw:
        return {}
    parsed = json.loads(raw)
    return parsed if isinstance(parsed, dict) else {}


def _try_import_pandas_ta():
    try:
        import pandas_ta_classic as ta  # type: ignore
    except Exception as exc:
        raise RuntimeError(f"加载 pandas-ta-classic 失败: {exc}") from exc
    return ta


def _series_payload(series: dict[str, Any]) -> dict[str, pd.Series]:
    """兼容旧调用点，实际逻辑下沉到独立模块，便于复用和收敛长度/索引策略。"""
    return build_series_payload(series)


def _normalize_group_name(raw_group: str) -> str:
    return GROUP_NAME_MAPPING.get(raw_group, raw_group.strip().lower().replace(" ", "_"))


def _group_display_name(raw_group: str) -> str:
    return GROUP_LABEL_MAPPING.get(raw_group, raw_group.replace("_", " ").title())


def _build_description(function_key: str, group_name: str, summary: str, usage: str) -> str:
    translated = summary.strip()
    for english, chinese in sorted(SUMMARY_TRANSLATION_MAP, key=lambda item: len(item[0]), reverse=True):
        translated = translated.replace(english, chinese)
    detail = FUNCTION_DETAIL_OVERRIDES.get(function_key)
    if not detail:
        purpose = GROUP_PURPOSE_MAP.get(group_name, "用于技术分析")
        detail = f"{translated or function_key}，主要用于{purpose}。"
    formula_hint = FUNCTION_FORMULA_HINTS.get(function_key, "")
    parts = [f"功能说明：{detail}"]
    if formula_hint:
        parts.append(f"计算公式：{formula_hint}")
    elif usage:
        parts.append(f"计算公式：{usage.strip()}")
    return "；".join(parts)


def _coerce_numeric(value: Any) -> float | int | None:
    if value is None:
        return None
    if isinstance(value, (pd.NA.__class__,)):  # pragma: no cover
        return None
    if isinstance(value, bool):
        return int(value)
    if isinstance(value, (float, int)):
        number = float(value)
    else:
        try:
            number = float(value)
        except (TypeError, ValueError):
            return None
    if not math.isfinite(number):
        return None
    rounded = round(number, 6)
    if rounded.is_integer():
        return int(rounded)
    return rounded


def _serialize_scalar(value: Any) -> float | int | str | None:
    numeric = _coerce_numeric(value)
    if numeric is not None:
        return numeric
    if isinstance(value, (str, bytes)):
        return value.decode("utf-8", errors="replace") if isinstance(value, bytes) else value
    if isinstance(value, (pd.Timestamp,)):
        return value.isoformat()
    return str(value)


def _serialize_value(value: Any) -> Any:
    if value is None:
        return None
    if isinstance(value, pd.DataFrame):
        return _serialize_frame(value)
    if isinstance(value, pd.Series):
        return _serialize_series(value)
    if isinstance(value, dict):
        return {str(key): _serialize_value(item) for key, item in value.items()}
    if isinstance(value, tuple):
        return [_serialize_value(item) for item in value]
    if isinstance(value, Iterable) and not isinstance(value, (str, bytes)):
        return [_serialize_value(item) for item in list(value)]
    return _serialize_scalar(value)


def _safe_list(value: Any) -> list[Any]:
    if value is None:
        return []
    if isinstance(value, pd.DataFrame):
        raise TypeError("DataFrame 需要按列拆分处理")
    if isinstance(value, pd.Series):
        return _serialize_series(value)
    if isinstance(value, Iterable) and not isinstance(value, (str, bytes, dict)):
        return [_serialize_value(item) for item in list(value)]
    return [_serialize_value(value)]


def _serialize_series(values: pd.Series) -> list[float | int | str | None]:
    return [_serialize_scalar(item) for item in values.tolist()]


def _serialize_frame(frame: pd.DataFrame) -> dict[str, list[float | int | str | None]]:
    return {str(column): _serialize_series(frame[column]) for column in frame.columns}


def _normalize_output_key(function_key: str, raw_key: str) -> str:
    key = str(raw_key).strip()
    upper = key.upper()
    if function_key == "MACD":
        if "H" in upper and "MACD" in upper:
            return "hist"
        if "S" in upper and "MACD" in upper:
            return "signal"
        return "macd"
    if function_key == "MACDEXT":
        if "H" in upper and "MACD" in upper:
            return "hist"
        if "S" in upper and "MACD" in upper:
            return "signal"
        return "macd"
    if function_key == "MACDFIX":
        if "H" in upper and "MACD" in upper:
            return "hist"
        if "S" in upper and "MACD" in upper:
            return "signal"
        return "macd"
    if upper.endswith("H") and "MACD" in upper:
        return "hist"
    if upper.endswith("S") and "MACD" in upper:
        return "signal"
    return key.lower()


def _serialize_tuple_result(items: tuple[Any, ...]) -> dict[str, Any] | list[Any]:
    if not items:
        return []
    if all(isinstance(item, pd.DataFrame) for item in items):
        merged: dict[str, Any] = {}
        for index, item in enumerate(items, start=1):
            assert isinstance(item, pd.DataFrame)
            for column in item.columns:
                merged[f"{index}_{column}"] = _serialize_series(item[column])
        return merged
    if all(isinstance(item, pd.Series) for item in items):
        return {f"value_{index}": _serialize_series(item) for index, item in enumerate(items, start=1)}
    return [_serialize_value(item) for item in items]


def _output_names_from_values(values: Any) -> list[str]:
    if isinstance(values, dict):
        return list(values.keys())
    return ["values"]


def _pattern_names() -> list[str]:
    ta = _try_import_pandas_ta()
    names: list[str] = []
    for key, value in getattr(ta, "Category", {}).items():
        if str(key).lower() == "candles" and isinstance(value, list):
            for item in value:
                name = str(item).strip()
                if not name:
                    continue
                if name.lower() == "cdl_pattern":
                    continue
                names.append(name)
    return sorted(dict.fromkeys(names))


def _is_pattern_function(function_key: str) -> bool:
    return function_key.startswith("CDL") and function_key not in {"CDL_Z"}


def _is_candle_transform_function(function_key: str) -> bool:
    return function_key in {"CDL_Z", "HA"}


def _pattern_output_names(function_key: str) -> list[str]:
    return ["signals"] if function_key == "CDL_PATTERN" else ["signal"]


def _derive_render_pane(function_key: str, group_name: str) -> str:
    if _is_candle_transform_function(function_key):
        return "sub"
    if _is_pattern_function(function_key):
        return "main"
    if function_key in {"BBANDS", "DONCHIAN", "KC", "ICHIMOKU", "ADX", "ADXR", "AROON", "STOCH", "STOCHRSI", "HT_SINE", "HT_PHASOR"}:
        return "main" if function_key in {"BBANDS", "DONCHIAN", "KC", "ICHIMOKU", "ADX", "ADXR", "AROON"} else "sub"
    return "main" if group_name in {"overlap_studies", "trend_indicators", "volatility_indicators"} else "sub"


def _normalize_parameter_name(name: str) -> str:
    return name.strip().lower().replace("-", "_")


def _normalized_parameters(parameters: dict[str, Any]) -> dict[str, Any]:
    out: dict[str, Any] = {}
    for key, value in parameters.items():
        out[_normalize_parameter_name(str(key))] = value
    return out


def _resolve_parameter_value(name: str, parameters: dict[str, Any]) -> Any:
    aliases = {
        "length": ["timeperiod", "length"],
        "fast": ["fast", "fastperiod"],
        "slow": ["slow", "slowperiod"],
        "mamode": ["mamode", "matype"],
        "periods": ["periods"],
        "minperiod": ["minperiod"],
        "maxperiod": ["maxperiod"],
        "lensig": ["lensig", "signalperiod"],
        "scalar": ["scalar"],
        "std": ["std", "nbdev"],
        "ddof": ["ddof"],
        "drift": ["drift"],
        "offset": ["offset"],
        "talib": ["talib"],
        "name": ["name"],
    }
    normalized = _normalized_parameters(parameters)
    keys = aliases.get(name, [name])
    for key in keys:
        if key in normalized:
            return normalized[key]
    return inspect._empty


def _resolve_series_argument(name: str, series: dict[str, pd.Series]) -> Any:
    aliases = {
        "open": ["open"],
        "open_": ["open"],
        "high": ["high"],
        "low": ["low"],
        "close": ["close"],
        "benchmark": ["benchmark", "close"],
        "fast": ["fast"],
        "slow": ["slow", "benchmark"],
        "trend": ["trend", "close"],
        "signal": ["signal", "close"],
        "volume": ["volume"],
        "periods": ["periods"],
    }
    keys = aliases.get(name, [name])
    for key in keys:
        if key in series:
            return series[key]
    return inspect._empty


def _is_series_annotation(annotation: Any) -> bool:
    if annotation is inspect._empty:
        return False
    if annotation is pd.Series:
        return True
    origin = get_origin(annotation)
    args = get_args(annotation)
    if origin is None:
        return getattr(annotation, "__name__", "") == "Series"
    return any(_is_series_annotation(arg) for arg in args)


def _call_by_signature(fn: Any, series: dict[str, pd.Series], parameters: dict[str, Any]) -> Any:
    signature = inspect.signature(fn)
    kwargs: dict[str, Any] = {}
    function_defaults = FUNCTION_PARAMETER_DEFAULTS.get(fn.__name__.upper(), {})
    for name, param in signature.parameters.items():
        if name in {"self", "cls"}:
            continue
        if _is_series_annotation(param.annotation):
            series_value = _resolve_series_argument(name, series)
            if series_value is not inspect._empty:
                kwargs[name] = series_value
                continue
        parameter_value = _resolve_parameter_value(name, parameters)
        if parameter_value is not inspect._empty:
            kwargs[name] = parameter_value
            continue
        if name in function_defaults:
            kwargs[name] = function_defaults[name]
            continue
        if param.default is inspect._empty and param.kind not in {inspect.Parameter.VAR_POSITIONAL, inspect.Parameter.VAR_KEYWORD}:
            raise ValueError(f"缺少函数 {fn.__name__} 所需参数：{name}")
    return fn(**kwargs)


def _infer_input_requirements(fn: Any, function_key: str, raw_group: str) -> list[str]:
    override = FUNCTION_INPUT_REQUIREMENTS.get(function_key)
    if override is not None:
        return override
    try:
        signature = inspect.signature(fn)
    except Exception:
        return ["open", "high", "low", "close"] if raw_group == "candles" else ["close"]
    requirements: list[str] = []
    for name, param in signature.parameters.items():
        if name in {"self", "cls", "talib", "offset", "ddof", "scalar", "length", "k", "d", "smooth_k", "mamode", "std", "c"}:
            continue
        if param.kind in {inspect.Parameter.VAR_POSITIONAL, inspect.Parameter.VAR_KEYWORD}:
            continue
        if _is_series_annotation(param.annotation):
            requirement = SIGNATURE_INPUT_ALIASES.get(name, name)
            requirements.append(requirement)
    if requirements:
        return list(dict.fromkeys(requirements))
    if {"high", "low", "close", "volume"} <= set(signature.parameters.keys()):
        return ["high", "low", "close", "volume"]
    if {"open", "high", "low", "close"} <= set(signature.parameters.keys()):
        return ["open", "high", "low", "close"]
    if {"high", "low", "close"} <= set(signature.parameters.keys()):
        return ["high", "low", "close"]
    if {"high", "low"} <= set(signature.parameters.keys()):
        return ["high", "low"]
    if "volume" in signature.parameters:
        return ["volume"]
    return ["open", "high", "low", "close"] if raw_group == "candles" else ["close"]


def _build_catalog() -> list[dict[str, Any]]:
    ta = _try_import_pandas_ta()
    category = getattr(ta, "Category", {})
    items: list[dict[str, Any]] = []
    for raw_group, keys in category.items():
        if not isinstance(keys, list):
            continue
        group_name = _normalize_group_name(str(raw_group))
        for raw_key in keys:
            function_key = str(raw_key).upper()
            render_pane = _derive_render_pane(function_key, group_name)
            fn = getattr(ta, raw_key, None)
            display_name = function_key
            summary = ""
            usage = ""
            if fn is not None and getattr(fn, "__doc__", None):
                summary = (fn.__doc__ or "").strip().splitlines()[0].strip()
                usage = "根据输入序列与参数调用 pandas-ta-classic 原生实现。"
            description = _build_description(function_key, group_name, summary, usage)
            output_names = ["values"]
            output_type = "series"
            if raw_group == "candles":
                output_names = _pattern_output_names(function_key)
                output_type = "object"
            items.append({
                "function_key": function_key,
                "display_name": display_name,
                "group_name": group_name,
                "description_zh": description,
                "function_type": "pattern" if raw_group == "candles" else "indicator",
                "parameters_schema": {},
                "input_series_requirements": _infer_input_requirements(fn, function_key, raw_group) if fn is not None else (
                    ["open", "high", "low", "close"] if raw_group == "candles" else ["close"]
                ),
                "output_schema": {
                    "type": output_type,
                    "output_names": output_names,
                    "render_pane": render_pane,
                },
                "render_schema": {
                    "main_series": [] if render_pane == "sub" else ["values"],
                    "sub_series": [] if render_pane == "main" else ["values"],
                    "markers": ["markers"] if raw_group == "candles" else [],
                    "events": [],
                },
                "warmup_bars": 0,
                "tags": [group_name],
            })
    items.sort(key=lambda item: (item["group_name"], item["function_key"]))
    return items


def _build_compute_result(function_key: str, raw_result: Any, warmup_bars: int, render_pane: str, series: dict[str, pd.Series]) -> dict[str, Any]:
    values: Any
    if isinstance(raw_result, pd.DataFrame):
        values = _serialize_frame(raw_result)
    elif isinstance(raw_result, pd.Series):
        values = _serialize_series(raw_result)
    elif isinstance(raw_result, tuple):
        values = _serialize_tuple_result(raw_result)
    elif isinstance(raw_result, dict):
        values = _serialize_value(raw_result)
    else:
        values = _safe_list(raw_result)
    render_payload: dict[str, Any] = {"main_series": [], "sub_series": [], "markers": [], "events": []}
    if isinstance(values, dict):
        for key, series_values in values.items():
            normalized_key = _normalize_output_key(function_key, key)
            is_pattern = _is_pattern_function(function_key)
            target_key = "main_series" if render_pane == "main" and not is_pattern else "sub_series"
            render_payload[target_key].append({
                "key": normalized_key,
                "name": normalized_key.upper() if normalized_key in {"macd", "signal", "hist"} else str(key),
                "pane": "sub" if is_pattern else render_pane,
                "style": "histogram" if is_pattern or normalized_key == "hist" else "line",
                "values": series_values,
            })
            if is_pattern:
                close_series = series.get("close")
                for index, value in enumerate(series_values):
                    numeric_value = _coerce_numeric(value)
                    if numeric_value in (None, 0):
                        continue
                    if close_series is None or index >= len(close_series):
                        continue
                    close_value = _serialize_scalar(close_series.iloc[index])
                    render_payload["markers"].append({
                        "index": index,
                        "time": index,
                        "value": close_value,
                        "label": str(key),
                        "color": "#22c55e" if float(numeric_value) > 0 else "#ef4444",
                        "side": "above" if float(numeric_value) > 0 else "below",
                        "description": f"{key} 命中",
                        "signal": numeric_value,
                    })
    else:
        target_key = "main_series" if render_pane == "main" and not _is_pattern_function(function_key) else "sub_series"
        single_key = _normalize_output_key(function_key, "values")
        render_payload[target_key].append({
            "key": single_key,
            "name": single_key.upper() if single_key in {"macd", "signal", "hist"} else function_key,
            "pane": "sub" if _is_pattern_function(function_key) else render_pane,
            "style": "histogram" if _is_pattern_function(function_key) or single_key == "hist" else "line",
            "values": values,
        })
    return {
        "function_key": function_key,
        "summary": f"{function_key} 已计算完成。",
        "warnings": [],
        "values": values,
        "render_payload": render_payload,
        "source_meta": {
            "output_names": _output_names_from_values(values),
            "render_pane": render_pane,
            "warmup_bars": warmup_bars,
        },
    }


def _call_function(ta: Any, function_key: str, series: dict[str, pd.Series], parameters: dict[str, Any]) -> tuple[Any, int, str]:
    key = function_key.strip().upper()
    if not key:
        raise ValueError("function_key 不能为空")

    if key == "ATR":
        result = ta.atr(high=series["high"], low=series["low"], close=series["close"], length=int(parameters.get("timeperiod", 14)))
        return result, 14, "sub"
    if key == "RSI":
        result = ta.rsi(series["close"], length=int(parameters.get("timeperiod", 14)))
        return result, 14, "sub"
    if key == "MAVP":
        result = ta.mavp(series["close"], periods=series.get("periods"), minperiod=int(parameters.get("minperiod", 2)), maxperiod=int(parameters.get("maxperiod", 30)), mamode=int(parameters.get("matype", 0)))
        return result, 30, "sub"
    if key in {"ACOS", "ASIN", "ATAN", "COSH", "EXP", "SINH", "TANH"}:
        fn = getattr(ta, key.lower(), None)
        if fn is None:
            raise ValueError(f"pandas-ta-classic 不支持函数 {key}")
        result = fn(series["close"])
        return result, 0, "sub"
    if key == "CDL_Z":
        result = ta.cdl_z(
            open_=series["open"],
            high=series["high"],
            low=series["low"],
            close=series["close"],
            length=int(parameters.get("length", 30)),
            full=parameters.get("full"),
            ddof=parameters.get("ddof"),
            offset=parameters.get("offset"),
        )
        return result, int(parameters.get("length", 30) or 30), "sub"
    if key == "HA":
        result = ta.ha(
            open_=series["open"],
            high=series["high"],
            low=series["low"],
            close=series["close"],
            offset=parameters.get("offset"),
        )
        return result, 0, "sub"
    if key.startswith("CDL_") or key.startswith("CDL"):
        pattern_name = key.replace("CDL_", "").replace("CDL", "").lower().lstrip("_")
        if pattern_name == "pattern":
            result = ta.cdl_pattern(open_=series["open"], high=series["high"], low=series["low"], close=series["close"], name="all")
            return result, 0, "main"
        fn = getattr(ta, f"cdl_{pattern_name}", None)
        if fn is None:
            result = ta.cdl_pattern(open_=series["open"], high=series["high"], low=series["low"], close=series["close"], name=pattern_name)
            return result, 0, "main"
        result = _call_by_signature(fn, series, parameters)
        return result, 0, "main"
    fn = getattr(ta, key.lower(), None)
    if fn is None:
        raise ValueError(f"暂不支持的 function_key：{key}")
    result = _call_by_signature(fn, series, parameters)
    render_pane = _derive_render_pane(key, _group_name_for_key(key))
    warmup_bars = int(parameters.get("timeperiod", parameters.get("length", parameters.get("fast", 0))) or 0)
    return result, warmup_bars, render_pane


def _group_name_for_key(function_key: str) -> str:
    key = function_key.strip().upper()
    if _is_candle_transform_function(key):
        return "candle_patterns"
    if _is_pattern_function(key):
        return "candle_patterns"
    if key in {"BBANDS", "DONCHIAN", "KC", "ICHIMOKU"}:
        return "volatility_indicators"
    if key in {"ADX", "ADXR", "AROON", "HT_DCPERIOD", "HT_DCPHASE", "HT_PHASOR", "HT_SINE", "HT_TRENDMODE", "HT_TRENDLINE"}:
        return "cycle_indicators" if key.startswith("HT_") else "trend_indicators"
    return "momentum_indicators"


def command_catalog() -> None:
    print(json.dumps({"items": _build_catalog()}, ensure_ascii=False))


def command_compute() -> None:
    ta = _try_import_pandas_ta()
    payload = _read_stdin_json()
    function_key = str(payload.get("function_key", "")).strip().upper()
    parameters = payload.get("parameters") if isinstance(payload.get("parameters"), dict) else {}
    series = payload.get("series") if isinstance(payload.get("series"), dict) else {}
    normalized_series = _series_payload(series)
    result, warmup_bars, render_pane = _call_function(ta, function_key, normalized_series, parameters)
    output = _build_compute_result(function_key, result, warmup_bars, render_pane, normalized_series)
    output["parameters"] = parameters
    output["series_meta"] = {
        "fields": sorted(normalized_series.keys()),
        "length": len(next(iter(normalized_series.values()))) if normalized_series else 0,
    }
    print(json.dumps(output, ensure_ascii=False))


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit("usage: pandas_ta_bridge.py <catalog|compute>")
    command = sys.argv[1].strip().lower()
    if command == "catalog":
        command_catalog()
        return
    if command == "compute":
        command_compute()
        return
    raise SystemExit(f"unknown command: {command}")


if __name__ == "__main__":
    main()
