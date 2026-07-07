# Директории с исходными файлами
AS_IS_DIR := as-is
TO_BE_DIR := to-be

# Исходные файлы .idef0
AS_IS_FILES := $(wildcard $(AS_IS_DIR)/*.idef0)
TO_BE_FILES := $(wildcard $(TO_BE_DIR)/*.idef0)

# Целевые файлы .svg (заменяем расширение)
AS_IS_SVG := $(AS_IS_FILES:.idef0=.svg)
TO_BE_SVG := $(TO_BE_FILES:.idef0=.svg)

# Все SVG файлы
ALL_SVG := $(AS_IS_SVG) $(TO_BE_SVG)

# Цель по умолчанию
all: $(ALL_SVG)

# Правило для генерации SVG из IDEF0
%.svg: %.idef0
	@echo "Генерация $@ из $<"
	@idef0-schematic < $< > $@

# Очистка сгенерированных файлов
clean:
	@echo "Удаление сгенерированных SVG файлов"
	@rm -f $(ALL_SVG)

# Перегенерация всех файлов
rebuild: clean all

.PHONY: all clean rebuild
