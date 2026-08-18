import csv
import sys


def escape_typst(s):
    """Escape special Typst characters in content blocks."""
    # Escape backslashes first
    s = s.replace("\\", "\\\\")
    # Escape @ to avoid label references
    s = s.replace("@", "@ ")
    # Escape # to avoid code evaluation
    s = s.replace("#", r"\#")
    # Replace semicolons with Typst linebreaks for list display
    s = s.replace(";", " \\\\ ")
    return s


def main():
    csv_path = sys.argv[1]
    output_path = sys.argv[2]

    with open(csv_path, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f, delimiter=";")
        rows = list(reader)

    lines = []
    lines.append('#import "../report-template/template.typ": lab_report')
    lines.append("")
    lines.append("#show: body => lab_report(")
    lines.append('  discipline: "Тестирование программного обеспечения",')
    lines.append('  lab_number: "1",')
    lines.append('  lab_title: "Программа и методика испытаний",')
    lines.append('  student_group: "ДИПРБ-31/2",')
    lines.append('  student_name: "Наследсков М.В.",')
    lines.append('  teacher_name: "Лаптев В.В.",')
    lines.append('  city: "Астрахань",')
    lines.append('  year: "2026",')
    lines.append("  body,")
    lines.append(")")
    lines.append("")
    lines.append("= Объект тестирования")
    lines.append("")
    lines.append(
        "Объектом тестирования является автоматизированная система «Центр разработчиков» — система управления процессом публикации и сопровождения веб-игр на платформе Welwise Games."
    )
    lines.append("")
    lines.append("Система состоит из следующих подсистем:")
    lines.append("+ подсистема авторизации и управления учётными записями;")
    lines.append("+ подсистема управления проектами;")
    lines.append("+ подсистема развёртывания;")
    lines.append("+ подсистема модерации;")
    lines.append("+ подсистема оркестрации серверов;")
    lines.append("+ подсистема аналитики и отчётности.")
    lines.append("")

    for i, row in enumerate(rows, 1):
        tid = escape_typst(row["Идентификатор"])
        title = escape_typst(row["Заголовок"])
        priority = escape_typst(row["Приоритет"])
        preconditions = escape_typst(row["Предусловия"])
        steps = escape_typst(row["Шаги выполнения"])
        expected = escape_typst(row["Ожидаемый результат"])
        test_data = escape_typst(row["Тестовые данные"])
        postconditions = escape_typst(row["Постусловия"])

        lines.append(f"== Тестовый сценарий — {title}")
        lines.append("")
        lines.append("#table(")
        lines.append("  columns: (2.5cm, 1fr),")
        lines.append("  stroke: 0.5pt,")
        lines.append("  inset: 6pt,")
        lines.append("  align: (center, left),")
        lines.append("  table.header([*Атрибут*], [*Значение*]),")
        lines.append(f"  [Идентификатор], [{tid}],")
        lines.append(f"  [Приоритет], [{priority}],")
        lines.append(f"  [Предусловия], [{preconditions}],")
        lines.append(f"  [Шаги выполнения], [{steps}],")
        lines.append(f"  [Ожидаемый результат], [{expected}],")
        lines.append(f"  [Тестовые данные], [{test_data}],")
        lines.append(f"  [Постусловия], [{postconditions}]")
        lines.append(")")
        lines.append("")

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    print(f"Generated {output_path} with {len(rows)} test cases")


if __name__ == "__main__":
    main()
