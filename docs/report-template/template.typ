// ============================================================
// lab-template.typ — Шаблон отчёта по лабораторной работе АГТУ
// ============================================================

#let lab_report(
  // Логотип
  logo_path: "logo.png",

  // Шапка университета
  university: [
    Федеральное агентство по рыболовству \
    Федеральное государственное бюджетное образовательное \
    учреждение высшего образования \
    «Астраханский государственный технический университет»
  ],
  university_note: [
    Система менеджмента качества в области образования, воспитания, науки и инноваций сертифицирована \
    ООО «ДКС РУС» по международному стандарту ISO 9001:2015
  ],

  // Сведения
  institute: "Информационных технологий и коммуникаций",
  direction: "09.03.04 Программная инженерия",
  profile: "Разработка программно-информационных систем",
  department: "«Автоматизированные системы обработки информации и управления»",

  // Работа
  discipline: "СУБД PostgreSQL",
  lab_number: "1",
  lab_title: "ТЕМА",

  // Студент
  student_group: "ДИПРБ-31",
  student_name: "Пушкин А.С.",

  // Проверяющий
  teacher_post: "ст. преподаватель",
  teacher_name: "Куркурин Н.Д.",
  teacher_extra: "(ученая степень, ученое звание, Фамилия И.О.)",

  // Город и год
  city: "АСТРАХАНЬ",
  year: "2026",

  // Основной контент
  body,
) = {
  // ============================================================
  // Глобальные настройки документа
  // ============================================================
  set page(
    paper: "a4",
    margin: (top: 1.8cm, bottom: 2cm, left: 2cm, right: 1.5cm),
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

    // Верхняя шапка
    #grid(
      columns: (auto, 1fr),
      gutter: 0.8cm,
      align: (center + horizon, center + horizon),

      image(logo_path, height: 2.6cm),
      [
        #align(center)[
          #set par(leading: 0.25em, spacing: 0pt)
          #text(weight: "bold", style: "italic", size: 12pt)[#university]
          \
          #text(weight: "bold", size: 6pt)[#university_note]
        ]
      ]
    )

    #v(1.1cm)

    // Блок с информацией об институте и т.д.
    #align(left)[
      #set par(leading: 0.3em)
      #grid(
        columns: (3.1cm, 1fr),
        row-gutter: 0.4em,
        align: (left, left),
        [Институт], underline(institute),
        [Направление], underline(direction),
        [Профиль], underline(profile),
        [Кафедра], underline(department)
      )
    ]

    #v(3.0cm)

    // Центральный заголовок
    #align(center)[
      // Правка 1: Используем жесткую сетку, чтобы идеально прижать строки друг к другу
      #grid(
        columns: 1fr,
        row-gutter: 1.2em, // Очень компактный интервал
        text(weight: "bold", size: 20pt)[Лабораторная работа № #lab_number],
        text(weight: "bold", size: 16pt)[#underline[«#lab_title»]],
        text(size: 14pt)[по дисциплине «#discipline»]
      )
    ]

    #v(1.8cm)

    // Таблица с исполнителем и преподавателем
    // Правка 2: Сделали правый блок уже (теперь колонки 50% на 50%)
    #table(
      columns: (50%, 50%),
      inset: 0.22cm,
      stroke: 0.5pt,

      [], [
        #set align(left)
        #set par(leading: 0.3em, spacing: 0.3em)
        #set text(size: 12pt) // Жестко фиксируем 12-й размер шрифта

        Работа выполнена студентом группы #student_group
        #v(0.3em)
        #grid(
          columns: (1fr, 1fr),
          gutter: 8pt,
          align: center,
          [
            // Правка 3: ФИО студента теперь выровнено по левому краю
            #box(width: 100%, stroke: (bottom: 0.5pt), outset: (bottom: 2pt))[
              #align(left)[#student_name]
            ] \
            #v(0.1em)
            #text(size: 8pt)[(Фамилия И.О.)]
          ],
          [
            #box(width: 100%, stroke: (bottom: 0.5pt), outset: (bottom: 2pt))[ ] \
            #v(0.1em)
            #text(size: 8pt)[подпись]
          ]
        )
      ],

      [], [
        #set align(left)
        #set par(leading: 0.3em, spacing: 0.3em)
        #set text(size: 12pt)

        Проверил работу:
        #v(0.3em)
        #align(center)[
          #box(width: 100%, stroke: (bottom: 0.5pt), outset: (bottom: 2pt))[
            #align(left)[#teacher_post #teacher_name]
          ] \
          #v(0.1em)
          #text(size: 8pt)[#teacher_extra]
        ]
      ],
    )

    // Блок "Работа защищена"
    #grid(
      columns: (50%, 50%), // Синхронизировано с таблицей
      [],
      [
        #align(center)[
          #box(align(left)[
            Работа защищена \
            «`___`» #box(width: 3cm, stroke: (bottom: 0.5pt)) #year г.
          ])
        ]
      ]
    )

    #v(1fr)

    #align(center)[
      #text(weight: "bold", size: 13pt)[#city – #year]
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
    v(0.25cm)
  }

  show heading.where(level: 2): it => {
    v(0.35cm)
    block(text(weight: "bold", size: 12pt, it))
    v(0.2cm)
  }

  // ============================================================
  // ОСНОВНОЙ КОНТЕНТ
  // ============================================================
  body
}
