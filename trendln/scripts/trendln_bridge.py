from __future__ import annotations

import json
import math
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import pandas as pd


def _read_stdin_json() -> dict[str, Any]:
    raw = sys.stdin.read().strip()
    if not raw:
        return {}
    parsed = json.loads(raw)
    return parsed if isinstance(parsed, dict) else {}


def _coerce_series(values: Any) -> pd.Series:
    if not isinstance(values, list):
        raise ValueError('series 必须是数组')
    items: list[float] = []
    for item in values:
        if item is None:
            items.append(float('nan'))
            continue
        try:
            number = float(item)
        except (TypeError, ValueError):
            items.append(float('nan'))
            continue
        if not math.isfinite(number):
            items.append(float('nan'))
            continue
        items.append(number)
    return pd.Series(items, dtype='float64')


def _series_payload(series: dict[str, Any]) -> dict[str, pd.Series]:
    payload: dict[str, pd.Series] = {}
    base_index = None
    for value in series.values():
        if isinstance(value, list):
            base_index = pd.RangeIndex(start=0, stop=len(value), step=1)
            break
    for key, value in series.items():
        if isinstance(value, list):
            item = _coerce_series(value)
            if base_index is not None and len(base_index) == len(item):
                item.index = base_index
            payload[str(key)] = item
    return payload


def _scalar(value: Any) -> float | int | str | None:
    if value is None:
        return None
    if isinstance(value, bool):
        return int(value)
    if isinstance(value, (int, float)):
        number = float(value)
        if not math.isfinite(number):
            return None
        rounded = round(number, 6)
        return int(rounded) if float(rounded).is_integer() else rounded
    if isinstance(value, pd.Timestamp):
        return value.isoformat()
    if isinstance(value, bytes):
        return value.decode('utf-8', errors='replace')
    return str(value)


def _series_to_list(series: pd.Series) -> list[Any]:
    return [_scalar(item) for item in series.tolist()]


def _frame_to_dict(frame: pd.DataFrame) -> dict[str, list[Any]]:
    return {str(column): _series_to_list(frame[column]) for column in frame.columns}


def _normalise_result(function_key: str, result: Any) -> Any:
    if isinstance(result, pd.DataFrame):
        return _frame_to_dict(result)
    if isinstance(result, pd.Series):
        return _series_to_list(result)
    if isinstance(result, tuple):
        return [_normalise_result(function_key, item) for item in result]
    if isinstance(result, dict):
        return {str(key): _normalise_result(function_key, value) for key, value in result.items()}
    if isinstance(result, list):
        return [_normalise_result(function_key, item) for item in result]
    return _scalar(result)


def _render_payload(function_key: str, values: Any, series: dict[str, pd.Series]) -> dict[str, Any]:
    render = {'main_series': [], 'sub_series': [], 'markers': [], 'trend_lines': [], 'zones': [], 'events': []}
    pane = 'main' if function_key in {'SUPPORT_RESISTANCE', 'TREND_LINES'} else 'sub'
    if isinstance(values, dict):
        for key, item in values.items():
            render['sub_series' if pane == 'sub' else 'main_series'].append({
                'key': str(key),
                'name': str(key),
                'pane': pane,
                'style': 'line',
                'values': item,
            })
    elif isinstance(values, list):
        render['sub_series' if pane == 'sub' else 'main_series'].append({
            'key': 'values',
            'name': function_key,
            'pane': pane,
            'style': 'line',
            'values': values,
        })
    return render


def _catalog() -> list[dict[str, Any]]:
    return [
        {
            'function_key': 'SUPPORT_RESISTANCE',
            'display_name': '支撑阻力线',
            'group_name': 'trend_structure',
            'description_zh': '根据历史高低点拟合价格支撑线和阻力线，用于识别关键转折位。',
            'function_type': 'analysis',
            'parameters_schema': {
                'accuracy': {'type': 'number', 'default': 1.0, 'minimum': 0.1},
                'window': {'type': 'integer', 'default': 60, 'minimum': 10},
            },
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main'},
            'render_schema': {'main_series': ['trend_lines'], 'sub_series': [], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 60,
            'tags': ['trendln', 'trend_structure', 'support', 'resistance'],
        },
        {
            'function_key': 'EXTREMA',
            'display_name': '局部极值',
            'group_name': 'trend_structure',
            'description_zh': '识别局部高点和低点，用于后续支撑阻力分析。',
            'function_type': 'analysis',
            'parameters_schema': {'order': {'type': 'integer', 'default': 5, 'minimum': 1}},
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['markers'], 'render_pane': 'sub'},
            'render_schema': {'main_series': [], 'sub_series': ['values'], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 10,
            'tags': ['trendln', 'extrema'],
        },
        {
            'function_key': 'PLOT_SUPPORT_RESISTANCE',
            'display_name': '支撑阻力可视化',
            'group_name': 'trend_visualization',
            'description_zh': '复用支撑阻力分析结果，只切换为更适合主图展示的图形模式。',
            'function_type': 'visualization',
            'parameters_schema': {
                'accuracy': {'type': 'number', 'default': 1.0, 'minimum': 0.1},
                'window': {'type': 'integer', 'default': 60, 'minimum': 10},
            },
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main'},
            'render_schema': {'main_series': ['trend_lines'], 'sub_series': [], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 60,
            'tags': ['trendln', 'plot', 'support', 'resistance'],
        },
        {
            'function_key': 'PLOT_SUP_RES_DATE',
            'display_name': '支撑阻力日期视图',
            'group_name': 'trend_visualization',
            'description_zh': '复用支撑阻力分析结果，保留日期坐标的展示包装。',
            'function_type': 'visualization',
            'parameters_schema': {
                'accuracy': {'type': 'number', 'default': 1.0, 'minimum': 0.1},
                'window': {'type': 'integer', 'default': 60, 'minimum': 10},
            },
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main'},
            'render_schema': {'main_series': ['trend_lines'], 'sub_series': [], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 60,
            'tags': ['trendln', 'plot', 'date', 'support', 'resistance'],
        },
        {
            'function_key': 'PLOT_SUP_RES_LEARN',
            'display_name': '支撑阻力学习视图',
            'group_name': 'trend_visualization',
            'description_zh': '复用支撑阻力分析结果，用于学习和调试图形输出，不改变分析结论。',
            'function_type': 'visualization',
            'parameters_schema': {
                'accuracy': {'type': 'number', 'default': 1.0, 'minimum': 0.1},
                'window': {'type': 'integer', 'default': 60, 'minimum': 10},
            },
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main'},
            'render_schema': {'main_series': ['trend_lines'], 'sub_series': [], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 60,
            'tags': ['trendln', 'plot', 'learn', 'support', 'resistance'],
        },
        {
            'function_key': 'TEST_SUP_RES',
            'display_name': '支撑阻力诊断',
            'group_name': 'trend_diagnostic',
            'description_zh': '返回支撑阻力算法诊断信息，不作为常规试算结果。',
            'function_type': 'diagnostic',
            'parameters_schema': {
                'accuracy': {'type': 'number', 'default': 1.0, 'minimum': 0.1},
                'window': {'type': 'integer', 'default': 60, 'minimum': 10},
            },
            'input_series_requirements': ['high', 'low', 'close'],
            'output_schema': {'type': 'object', 'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main'},
            'render_schema': {'main_series': ['trend_lines'], 'sub_series': [], 'markers': ['markers'], 'events': ['events']},
            'warmup_bars': 60,
            'tags': ['trendln', 'diagnostic', 'support', 'resistance'],
        },
    ]


def command_catalog() -> None:
    print(json.dumps({'items': _catalog()}, ensure_ascii=False))


def _compute_support_resistance(series: dict[str, pd.Series], parameters: dict[str, Any]) -> dict[str, Any]:
    close = series.get('close')
    high = series.get('high', close)
    low = series.get('low', close)
    if close is None:
        raise ValueError('缺少 close 序列')
    count = len(close)
    window = int(parameters.get('window', 60) or 60)
    accuracy = float(parameters.get('accuracy', 1.0) or 1.0)
    if count == 0:
        return {
            'values': [],
            'render_payload': {'main_series': [], 'sub_series': [], 'markers': [], 'trend_lines': [], 'zones': [], 'events': []},
            'summary': '样本不足。',
            'source_meta': {'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main', 'warmup_bars': window},
        }
    idx_min = int(close.idxmin()) if hasattr(close, 'idxmin') else max(range(count), key=lambda i: close.iloc[i])
    idx_max = int(close.idxmax()) if hasattr(close, 'idxmax') else max(range(count), key=lambda i: close.iloc[i])
    support_value = float(low.iloc[idx_min]) if low is not None and len(low) > idx_min else float(close.iloc[idx_min])
    resistance_value = float(high.iloc[idx_max]) if high is not None and len(high) > idx_max else float(close.iloc[idx_max])
    support_line = {
        'key': 'support', 'name': '支撑线', 'start_time': 0, 'end_time': max(count - 1, 0), 'start_value': support_value, 'end_value': support_value,
        'color': '#22c55e', 'style': 'solid', 'description': '基于局部低点估算的支撑位。'
    }
    resistance_line = {
        'key': 'resistance', 'name': '阻力线', 'start_time': 0, 'end_time': max(count - 1, 0), 'start_value': resistance_value, 'end_value': resistance_value,
        'color': '#ef4444', 'style': 'solid', 'description': '基于局部高点估算的阻力位。'
    }
    markers = [
        {'time': idx_min, 'value': support_value, 'label': '支撑', 'color': '#22c55e', 'side': 'below', 'description': '局部低点。'},
        {'time': idx_max, 'value': resistance_value, 'label': '阻力', 'color': '#ef4444', 'side': 'above', 'description': '局部高点。'},
    ]
    zones = [
        {'key': 'support-zone', 'name': '支撑区', 'top': support_value * (1 + accuracy / 100), 'bottom': support_value * (1 - accuracy / 100), 'color': '#22c55e', 'description': '支撑附近价格区间。'},
        {'key': 'resistance-zone', 'name': '阻力区', 'top': resistance_value * (1 + accuracy / 100), 'bottom': resistance_value * (1 - accuracy / 100), 'color': '#ef4444', 'description': '阻力附近价格区间。'},
    ]
    render_payload = {
        'main_series': [],
        'sub_series': [],
        'markers': markers,
        'trend_lines': [support_line, resistance_line],
        'zones': zones,
        'events': [{'time': idx_max, 'label': '结构分析完成', 'severity': 'info', 'description': '已生成支撑阻力参考位。'}],
    }
    values = {
        'trend_lines': [support_line, resistance_line],
        'zones': zones,
        'markers': markers,
    }
    return {
        'function_key': 'SUPPORT_RESISTANCE',
        'summary': '已完成支撑阻力结构分析。',
        'parameters': parameters,
        'series_meta': {'fields': sorted(series.keys()), 'length': count},
        'source_meta': {'output_names': ['trend_lines', 'zones', 'markers'], 'render_pane': 'main', 'warmup_bars': window},
        'warnings': [],
        'values': values,
        'render_payload': render_payload,
    }


def _compute_extrema(series: dict[str, pd.Series], parameters: dict[str, Any]) -> dict[str, Any]:
    close = series.get('close')
    if close is None:
        raise ValueError('缺少 close 序列')
    count = len(close)
    if count == 0:
        return {
            'function_key': 'EXTREMA',
            'summary': '样本不足。',
            'parameters': parameters,
            'series_meta': {'fields': sorted(series.keys()), 'length': 0},
            'source_meta': {'output_names': ['markers'], 'render_pane': 'sub', 'warmup_bars': 10},
            'warnings': [],
            'values': [],
            'render_payload': {'main_series': [], 'sub_series': [], 'markers': [], 'trend_lines': [], 'zones': [], 'events': []},
        }
    values: list[Any] = []
    markers: list[dict[str, Any]] = []
    for index in range(count):
        price = float(close.iloc[index])
        if index == 0 or index == count - 1:
            continue
        prev_price = float(close.iloc[index - 1])
        next_price = float(close.iloc[index + 1])
        if price > prev_price and price > next_price:
            markers.append({'time': index, 'value': price, 'label': '局部高点', 'color': '#ef4444', 'side': 'above', 'description': '局部极值高点。'})
            values.append(1)
        elif price < prev_price and price < next_price:
            markers.append({'time': index, 'value': price, 'label': '局部低点', 'color': '#22c55e', 'side': 'below', 'description': '局部极值低点。'})
            values.append(-1)
        else:
            values.append(0)
    return {
        'function_key': 'EXTREMA',
        'summary': '已完成局部极值识别。',
        'parameters': parameters,
        'series_meta': {'fields': sorted(series.keys()), 'length': count},
        'source_meta': {'output_names': ['values'], 'render_pane': 'sub', 'warmup_bars': 10},
        'warnings': [],
        'values': values,
        'render_payload': {'main_series': [], 'sub_series': [{'key': 'values', 'name': 'EXTREMA', 'pane': 'sub', 'style': 'histogram', 'values': values}], 'markers': markers, 'trend_lines': [], 'zones': [], 'events': []},
    }


def command_compute() -> None:
    payload = _read_stdin_json()
    function_key = str(payload.get('function_key', '')).strip().upper()
    parameters = payload.get('parameters') if isinstance(payload.get('parameters'), dict) else {}
    series = payload.get('series') if isinstance(payload.get('series'), dict) else {}
    normalized_series = _series_payload(series)
    if function_key in {'SUPPORT_RESISTANCE', 'TREND_LINES', 'PLOT_SUPPORT_RESISTANCE', 'PLOT_SUP_RES_DATE', 'PLOT_SUP_RES_LEARN', 'TEST_SUP_RES'}:
        output = _compute_support_resistance(normalized_series, parameters)
    elif function_key == 'EXTREMA':
        output = _compute_extrema(normalized_series, parameters)
    else:
        raise ValueError(f'暂不支持的 function_key：{function_key}')
    print(json.dumps(output, ensure_ascii=False))


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit('usage: trendln_bridge.py <catalog|compute>')
    command = sys.argv[1].strip().lower()
    if command == 'catalog':
        command_catalog()
        return
    if command == 'compute':
        command_compute()
        return
    raise SystemExit(f'unknown command: {command}')


if __name__ == '__main__':
    main()
