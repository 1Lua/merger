#!/bin/bash

# Название исполняемого файла
BINARY=merger
# Директория для установки
INSTALL_DIR=/usr/local/bin

cd build

# Проверяем, существует ли файл
if [ ! -f "$BINARY" ]; then
  echo "Ошибка: Файл '$BINARY' не найден. Сначала соберите проект."
  exit 1
fi

# Устанавливаем файл в системную директорию
sudo install -m 0755 "$BINARY" "$INSTALL_DIR"

# Проверяем успешность установки
if [ $? -eq 0 ]; then
  echo "Файл '$BINARY' успешно установлен в '$INSTALL_DIR'."
else
  echo "Ошибка установки файла '$BINARY'."
  exit 1
fi