// ============================================================
// lab-template.typ — Шаблон отчёта по лабораторной работе АГТУ
// ============================================================

#let lab_report(
  // Шапка университета
  logo_path: "logo.png",
  university: "Федеральное агентство по рыболовству \n Федеральное государственное бюджетное образовательное \n учреждение высшего образования \n «Астраханский государственный технический университет» \n Система менеджмента качества в области образования, воспитания, науки и инноваций сертифицирована \n ООО «ДКС РУС» по международному стандарту ISO 9001:2015",
  institute: "ИНСТИТУТ ИНФОРМАЦИОННЫХ ТЕХНОЛОГИЙ И КОММУНИКАЦИЙ",
  department: "КАФЕДРА АВТОМАТИЗИРОВАННЫХ СИСТЕМ ОБРАБОТКИ ИНФОРМАЦИИ И УПРАВЛЕНИЯ",

  // Дисциплина и работа
  discipline: "Наименование дисциплины",
  lab_number: "1",
  lab_title: "Наименование лабораторной работы",
  variant: "1",

  // Кто выполнил
  student_group: "Группа",
  student_name: "Фамилия И.О.",

  // Кто проверил
  teacher_name: "Фамилия И.О.",

  // Город и год
  city: "Астрахань",
  year: "2024",

  // Основной контент
  body,
) = {

  // ============================================================
  // Глобальные настройки документа
  // ============================================================
  set page(
    paper: "a4",
    margin: (top: 2cm, bottom: 2cm, left: 2cm, right: 1.5cm),
  )

  set text(
    font: "Times New Roman",
    size: 12pt,
    lang: "ru",
  )

  // ============================================================
  // ТИТУЛЬНЫЙ ЛИСТ
  // ============================================================
  page(numbering: none)[
    #set par(leading: 0.55em)
    #set align(center)

    // Шапка: Логотип слева, текст справа
    #grid(
      columns: (3.5cm, 1fr),
      gutter: 15pt,
      align: (center + horizon, center + horizon),
      image(logo_path, width: 100%),
      text(weight: "bold", style: "italic", size: 12pt)[#university]
    )

    #v(0.75cm)

    // Институт и кафедра
    #align(center)[
      #text(weight: "bold", size: 12pt)[#institute] \
      #text(weight: "bold", size: 12pt)[#department]
    ]

    #v(2.7cm)

    // Наименование дисциплины
    #align(center)[
      #text(weight: "bold", size: 12pt)[#underline(discipline)] \
      #text(style: "italic", size: 12pt)[(наименование дисциплины)]
    ]

    #v(1.5cm)

    // Блок ОТЧЕТ с названием работы
    #align(center)[
      #text(weight: "bold", size: 12pt)[ОТЧЕТ] \
      о выполнении индивидуального задания к лабораторной работе № #lab_number \
      «#lab_title» \
      Вариант № #variant
    ]

    #v(1fr)

    // Блок подписей (выровнен по правому краю)
    #grid(
      columns: (1fr, auto),
      [],
      [
        #set align(left)
        #set par(leading: 0.65em)

        Выполнил: \
        студент гр. #student_group \
        #underline(student_name) \
        « \_\_\_\_ » \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ #year г.

        #v(1.5em)

        Максимальное количество баллов \_\_\_\_\_\_ \
        ЗАЩИЩЕНО: \
        Получено баллов \_\_\_\_\_\_ \
        Преподаватель: #underline(teacher_name) \
        « \_\_\_\_ » \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ #year г.
      ]
    )

    #v(1fr)

    #align(center)[
      #text(weight: "bold")[#city – #year]
    ]
  ]

  // ============================================================
  // НАСТРОЙКИ ДЛЯ ОСНОВНОГО ТЕКСТА
  // ============================================================
  set page(numbering: "1", number-align: center)
  counter(page).update(1)

  set par(
    justify: true,
    leading: 1em,
    first-line-indent: 1.25cm,
  )

  set heading(numbering: "1.1")

  show heading.where(level: 1): it => {
    v(0.5cm)
    block(text(weight: "bold", size: 12pt, it))
    v(0.3cm)
    par(text(size: 0pt, h(0pt)))
    v(-1em)
  }

  show heading.where(level: 2): it => {
    v(0.4cm)
    block(text(weight: "bold", size: 12pt, it))
    v(0.2cm)
    par(text(size: 0pt, h(0pt)))
    v(-1em)
  }

  // ============================================================
  // ОСНОВНОЙ КОНТЕНТ
  // ============================================================
  body
}
