"""Уплотнённые параметры BPMN-генератора для схем в статье (узкая полоса набора)."""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'business-process', 'bpmn', 'src'))
import bpmn  # noqa: E402

bpmn.COLW, bpmn.ROWH, bpmn.TOPPAD = 100, 88, 22
bpmn.TASK_W, bpmn.TASK_H, bpmn.GW, bpmn.EV = 92, 72, 40, 30
bpmn.FONT = 'fontFamily=Times New Roman;fontSize=13;'
bpmn.FONT_SMALL = 'fontFamily=Times New Roman;fontSize=12;'

Diagram = bpmn.Diagram
