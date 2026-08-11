from __future__ import annotations

import json
import math
import re
import sys
from collections import OrderedDict
import os
from pathlib import Path
from typing import Any

import numpy as np


ROOT = Path(__file__).resolve().parents[2]
TA_LIB_PYTHON_ROOT = Path(
    os.getenv("TA_LIB_PYTHON_SOURCE_ROOT", str(ROOT / "ta-lib" / "ta-lib-python"))
).resolve()
TA_LIB_DOCS_ROOT = Path(
    os.getenv("TA_LIB_DOCS_ROOT", str(TA_LIB_PYTHON_ROOT / "docs" / "func_groups"))
).resolve()

if str(TA_LIB_PYTHON_ROOT) not in sys.path:
    sys.path.insert(0, str(TA_LIB_PYTHON_ROOT))

try:
    import talib
    from talib import abstract
except Exception as exc:  # pragma: no cover
    raise RuntimeError(f"加载 TA-Lib Python 运行时失败: {exc}") from exc


GROUP_NAME_MAPPING: dict[str, str] = {
    "Cycle Indicators": "cycle_indicators",
    "Math Operators": "math_operators",
    "Math Transform": "math_transform",
    "Momentum Indicators": "momentum_indicators",
    "Overlap Studies": "overlap_studies",
    "Pattern Recognition": "pattern_recognition",
    "Price Transform": "price_transform",
    "Statistic Functions": "statistic_functions",
    "Volatility Indicators": "volatility_indicators",
    "Volume Indicators": "volume_indicators",
}

MAIN_PANE_GROUPS = {
    "overlap_studies",
    "price_transform",
}

FUNCTION_DOCS: dict[str, dict[str, str]] = {}

# 常见英文摘要到中文摘要的映射，先覆盖核心术语，避免直接暴露英文说明。
SUMMARY_TRANSLATION_MAP: list[tuple[str, str]] = [
    ("Moving Average Convergence/Divergence Fix 12/26", "固定参数版移动平均收敛/发散指标 12/26"),
    ("Moving Average Convergence/Divergence", "移动平均收敛/发散指标"),
    ("Normalized Average True Range", "标准化平均真实波幅"),
    ("Average True Range", "平均真实波幅"),
    ("Relative Strength Index", "相对强弱指数"),
    ("Simple Moving Average", "简单移动平均"),
    ("Exponential Moving Average", "指数移动平均"),
    ("Double Exponential Moving Average", "双重指数移动平均"),
    ("Triple Exponential Moving Average", "三重指数移动平均"),
    ("Weighted Moving Average", "加权移动平均"),
    ("Kaufman Adaptive Moving Average", "考夫曼自适应移动平均"),
    ("Moving average with variable period", "可变周期移动平均"),
    ("Moving average", "移动平均"),
    ("Bollinger Bands", "布林带"),
    ("Hilbert Transform - Dominant Cycle Period", "希尔伯特变换 - 主导周期长度"),
    ("Hilbert Transform - Dominant Cycle Phase", "希尔伯特变换 - 主导周期相位"),
    ("Hilbert Transform - Phasor Components", "希尔伯特变换 - 相量分量"),
    ("Hilbert Transform - SineWave", "希尔伯特变换 - 正弦波"),
    ("Hilbert Transform - Trend vs Cycle Mode", "希尔伯特变换 - 趋势与周期模式"),
    ("Hilbert Transform - Instantaneous Trendline", "希尔伯特变换 - 瞬时趋势线"),
    ("Accelaration Bands", "加速带"),
    ("Acceleration Bands", "加速带"),
    ("Commodity Channel Index", "商品通道指数"),
    ("Chande Momentum Oscillator", "钱德动量摆动指标"),
    ("Balance Of Power", "力量平衡"),
    ("Money Flow Index", "资金流量指标"),
    ("Average Directional Movement Index Rating", "平均趋向指数评级"),
    ("Average Directional Movement Index", "平均趋向指数"),
    ("Directional Movement Index", "方向动量指标"),
    ("Directional Movement Index", "方向动量指标"),
    ("Trend vs Cycle Mode", "趋势与周期模式"),
    ("Phasor Components", "相量分量"),
    ("SineWave", "正弦波"),
    ("Triangle Moving Average", "三角移动平均"),
    ("Triangular Moving Average", "三角移动平均"),
    ("Acceleration", "加速度"),
    ("Rate-Of-Change", "变动率"),
]

FUNCTION_DETAIL_OVERRIDES: dict[str, str] = {
    "SMA": "对价格序列做等权平滑，适合过滤短期噪音并观察基础趋势。",
    "EMA": "对价格序列做指数加权平滑，更重视最近数据，适合跟踪趋势变化。",
    "RSI": "衡量价格上涨与下跌的相对强弱，常用于识别超买和超卖状态。",
    "ATR": "衡量价格波动幅度，不区分方向，常用于波动率判断和止损设置。",
    "NATR": "将平均真实波幅标准化到价格尺度，便于跨品种比较波动强弱。",
    "MACD": "通过快慢均线差、信号线和柱状图衡量趋势动量变化，适合观察趋势转折。",
    "BBANDS": "以中轨均线加上下轨标准差构建价格通道，适合观察波动收缩和扩张。",
    "HT_DCPERIOD": "估计当前市场主导周期长度，适合判断周期是否在拉长或缩短。",
    "HT_DCPHASE": "估计当前周期所处相位，适合观察周期推进阶段和可能的拐点。",
    "HT_PHASOR": "通过希尔伯特变换拆分同相和正交分量，适合分析周期相位与转向。",
    "HT_SINE": "输出周期正弦波及领先正弦波，常用于辅助观察周期拐点和节奏变化。",
    "HT_TRENDMODE": "判断当前更偏趋势还是周期状态，适合做市场状态切换判断。",
    "HT_TRENDLINE": "提取平滑的瞬时趋势线，适合观察价格主方向和噪音过滤后的轮廓。",
    "ADX": "衡量趋势强度而不区分方向，适合判断行情是否进入强趋势阶段。",
    "ADXR": "对 ADX 做进一步平滑，适合更稳定地评估趋势强度。",
    "TRANGE": "计算单根 K 线的真实波幅，是 ATR 等波动类指标的基础。",
    "CCI": "衡量价格相对均值的偏离程度，常用于识别超买超卖和动量变化。",
    "CMO": "通过涨跌净动量衡量市场强弱，适合观察动量转折。",
    "MFI": "结合价格和成交量衡量资金流入流出，常用于识别超买超卖。",
    "ROC": "衡量价格变化率，适合观察动量加速或减速。",
    "ROCP": "衡量相对变化百分比，适合做归一化动量比较。",
    "ROCR": "衡量价格相对前值的比例变化，适合观察趋势变化幅度。",
}

GROUP_PURPOSE_MAP: dict[str, str] = {
    "cycle_indicators": "分析市场周期长度、周期相位和趋势/周期切换",
    "overlap_studies": "对价格进行平滑、跟踪和通道分析，辅助观察趋势",
    "momentum_indicators": "衡量价格变化速度、强弱和动量转折",
    "volatility_indicators": "衡量价格波动幅度和波动强度",
    "volume_indicators": "结合成交量评估资金参与度和价格推动力",
    "price_transform": "把原始价格转换成更适合分析的价格序列",
    "statistic_functions": "做统计归纳、离散度和集中趋势分析",
    "math_transform": "对单序列进行数学变换，得到新的分析序列",
    "math_operators": "对多个序列做算术运算和组合变换",
    "pattern_recognition": "识别 K 线组合和形态信号",
}


def _read_stdin_json() -> dict[str, Any]:
    raw = sys.stdin.read().strip()
    if not raw:
        return {}
    parsed = json.loads(raw)
    return parsed if isinstance(parsed, dict) else {}


def _series_to_numpy(name: str, values: Any) -> np.ndarray:
    if not isinstance(values, list) or not values:
        raise ValueError(f"序列 {name} 不能为空")
    normalized: list[float] = []
    for item in values:
        if item is None:
            raise ValueError(f"序列 {name} 含有空值，暂不支持")
        number = float(item)
        if not math.isfinite(number):
            raise ValueError(f"序列 {name} 含有非法数值")
        normalized.append(number)
    return np.asarray(normalized, dtype=float)


def _flatten_required_inputs(input_names: OrderedDict[str, Any]) -> list[str]:
    required: list[str] = []
    for value in input_names.values():
        if isinstance(value, list):
            required.extend(str(item) for item in value)
        else:
            required.append(str(value))
    deduped: list[str] = []
    for item in required:
        if item not in deduped:
            deduped.append(item)
    return deduped


def _normalize_group_name(group_name: str) -> str:
    return GROUP_NAME_MAPPING.get(group_name, group_name.strip().lower().replace(" ", "_"))


def _group_display_name(group_name: str) -> str:
    reverse_mapping = {value: key for key, value in GROUP_NAME_MAPPING.items()}
    return reverse_mapping.get(group_name, group_name.replace("_", " ").title())


def _load_function_docs() -> dict[str, dict[str, str]]:
    group_dir = TA_LIB_DOCS_ROOT
    if not group_dir.exists():
        return {}
    docs: dict[str, dict[str, str]] = {}
    pattern = re.compile(
        r"### ([A-Z0-9_]+) - (.+?)\n(?:NOTE: The ``([A-Z0-9_]+)`` function has an unstable period\.\s+)?```python\n(.+?)\n```",
        re.S,
    )
    for path in sorted(group_dir.glob("*.md")):
        group_name = _group_display_name(path.stem)
        text = path.read_text(encoding="utf-8")
        for match in pattern.finditer(text):
            function_key, summary, note_key, usage = match.groups()
            docs[function_key] = {
                "summary": summary.strip(),
                "usage": usage.strip(),
                "note": "该函数存在不稳定周期。" if note_key else "",
                "group_name": group_name,
            }
    return docs


def _translate_summary_to_cn(summary: str) -> str:
    translated = summary.strip()
    for english, chinese in sorted(SUMMARY_TRANSLATION_MAP, key=lambda item: len(item[0]), reverse=True):
        translated = translated.replace(english, chinese)
    translated = translated.replace(" / ", " / ")
    return translated


def _build_function_description(function_key: str, summary: str, usage: str, note: str, group_name: str) -> str:
    summary_cn = _translate_summary_to_cn(summary)
    group_purpose = GROUP_PURPOSE_MAP.get(group_name, "用于技术分析")
    detail = FUNCTION_DETAIL_OVERRIDES.get(function_key)
    if not detail:
        detail = f"{summary_cn}，主要用于{group_purpose}。"
    parts = [f"功能说明：{detail}"]
    if usage:
        parts.append(f"计算公式：{usage}")
    if note:
        parts.append(f"注意：{note}")
    return "；".join(parts)


def _parameter_schema(parameters: OrderedDict[str, Any]) -> dict[str, Any]:
    schema: dict[str, Any] = {}
    for key, default in parameters.items():
        param_type = "string"
        if isinstance(default, bool):
            param_type = "boolean"
        elif isinstance(default, int) and not isinstance(default, bool):
            param_type = "integer"
        elif isinstance(default, float):
            param_type = "number"
        field: dict[str, Any] = {"type": param_type, "default": default}
        if key.endswith("period"):
            field["min"] = 1
        schema[str(key)] = field
    return schema


def _coerce_parameters(function: abstract.Function, raw_parameters: dict[str, Any]) -> dict[str, Any]:
    normalized: dict[str, Any] = {}
    for key, default in function.parameters.items():
        if key not in raw_parameters:
            normalized[key] = default
            continue
        value = raw_parameters[key]
        if isinstance(default, bool):
            normalized[key] = bool(value)
        elif isinstance(default, int) and not isinstance(default, bool):
            normalized[key] = int(value)
        elif isinstance(default, float):
            normalized[key] = float(value)
        else:
            normalized[key] = value
    for key, value in raw_parameters.items():
        if key not in normalized:
            normalized[key] = value
    return normalized


def _serialize_scalar(value: Any) -> float | int | None:
    if value is None:
        return None
    if isinstance(value, (np.floating, float)):
        if np.isnan(value):
            return None
        return round(float(value), 6)
    if isinstance(value, (np.integer, int)):
        return int(value)
    converted = float(value)
    if math.isnan(converted):
        return None
    rounded = round(converted, 6)
    if rounded.is_integer():
        return int(rounded)
    return rounded


def _serialize_output_array(values: Any) -> list[float | int | None]:
    array = np.asarray(values)
    if array.ndim != 1:
        raise ValueError("TA-Lib 输出维度异常")
    return [_serialize_scalar(item) for item in array.tolist()]


def _normalize_result_series(result: Any, output_count: int) -> list[Any]:
    if isinstance(result, tuple):
        return list(result)

    array = np.asarray(result)
    if output_count > 1 and array.ndim == 2:
        if array.shape[0] == output_count:
            return [array[index] for index in range(output_count)]
        if array.shape[1] == output_count:
            return [array[:, index] for index in range(output_count)]
    return [result]


def _build_result_values(function_key: str, output_names: list[str], result: Any) -> Any:
    outputs = _normalize_result_series(result, len(output_names))
    serialized = {
        output_names[index] if index < len(output_names) else f"output_{index + 1}": _serialize_output_array(values)
        for index, values in enumerate(outputs)
    }
    if len(serialized) == 1:
        return next(iter(serialized.values()))
    if function_key == "BBANDS":
        serialized["upper"] = serialized.get("upperband")
        serialized["middle"] = serialized.get("middleband")
        serialized["lower"] = serialized.get("lowerband")
    if function_key == "MACD":
        serialized["line"] = serialized.get("macd")
        serialized["signal"] = serialized.get("macdsignal")
        serialized["hist"] = serialized.get("macdhist")
    return serialized


def _build_catalog_item(function_key: str, group_name: str) -> dict[str, Any]:
    function = abstract.Function(function_key)
    normalized_group = _normalize_group_name(group_name)
    output_names = [str(item) for item in function.output_names]
    output_type = "series" if len(output_names) == 1 else "object"
    render_pane = "main" if normalized_group in MAIN_PANE_GROUPS else "sub"
    doc = FUNCTION_DOCS.get(function_key, {})
    summary = str(doc.get("summary", "")).strip()
    usage = str(doc.get("usage", "")).strip()
    note = str(doc.get("note", "")).strip()
    return {
        "function_key": function_key,
        "display_name": function_key,
        "group_name": normalized_group,
        "description_zh": _build_function_description(function_key, summary, usage, note, normalized_group),
        "parameters_schema": _parameter_schema(function.parameters),
        "input_series_requirements": _flatten_required_inputs(function.input_names),
        "output_schema": {
            "type": output_type,
            "output_names": output_names,
            "render_pane": render_pane,
        },
        "warmup_bars": int(getattr(function, "lookback", 0) or 0),
        "tags": [normalized_group],
    }


def handle_catalog() -> dict[str, Any]:
    global FUNCTION_DOCS
    if not FUNCTION_DOCS:
        FUNCTION_DOCS = _load_function_docs()
    items: list[dict[str, Any]] = []
    for group_name, function_keys in talib.get_function_groups().items():
        for function_key in function_keys:
            items.append(_build_catalog_item(str(function_key).upper(), str(group_name)))
    items.sort(key=lambda item: (item["group_name"], item["function_key"]))
    return {"items": items}


def handle_compute(payload: dict[str, Any]) -> dict[str, Any]:
    function_key = str(payload.get("function_key", "")).strip().upper()
    if not function_key:
        raise ValueError("function_key 不能为空")

    function = abstract.Function(function_key)
    raw_series = payload.get("series")
    if not isinstance(raw_series, dict):
        raise ValueError("series 不能为空")
    required_inputs = _flatten_required_inputs(function.input_names)
    series = {name: _series_to_numpy(name, raw_series.get(name)) for name in required_inputs}

    lengths = {name: int(values.shape[0]) for name, values in series.items()}
    unique_lengths = {length for length in lengths.values()}
    if len(unique_lengths) != 1:
        raise ValueError(f"输入序列长度不一致: {lengths}")

    parameters = _coerce_parameters(function, payload.get("parameters") if isinstance(payload.get("parameters"), dict) else {})
    result = function(series, **parameters)
    output_names = [str(item) for item in function.output_names]
    normalized_group = _normalize_group_name(
        next((name for name, keys in talib.get_function_groups().items() if function_key in keys), "other")
    )
    values = _build_result_values(function_key, output_names, result)
    input_length = next(iter(unique_lengths), 0)
    warmup = int(getattr(function, "lookback", 0) or 0)
    warnings: list[str] = []
    if input_length <= warmup:
        warnings.append("样本长度不足，前置值可能为空。")

    return {
        "function_key": function_key,
        "group_name": normalized_group,
        "summary": f"{function_key} 已计算完成。",
        "parameters": parameters,
        "series_meta": {
            "fields": required_inputs,
            "length": input_length,
        },
        "warnings": warnings,
        "values": values,
        "source_meta": {
            "warmup_bars": warmup,
            "output_names": output_names,
            "render_pane": "main" if normalized_group in MAIN_PANE_GROUPS else "sub",
        },
    }


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit("用法: talib_bridge.py <catalog|compute>")

    command = sys.argv[1].strip().lower()
    payload = _read_stdin_json()
    if command == "catalog":
        result = handle_catalog()
    elif command == "compute":
        result = handle_compute(payload)
    else:
        raise SystemExit(f"不支持的命令: {command}")

    sys.stdout.write(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:  # pragma: no cover
        sys.stderr.write(str(exc))
        sys.exit(1)
