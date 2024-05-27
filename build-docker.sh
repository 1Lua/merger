#!/bin/bash

# Проверяем, установлен ли Docker
if ! command -v docker &> /dev/null
then
    echo "Ошибка: Docker не установлен. Пожалуйста, установите Docker и попробуйте снова."
    exit 1
fi

# Название проекта и имя Docker образа
PROJECT_NAME=merger
DOCKER_IMAGE=merger-builder

# Запуск Docker для сборки проекта
docker build -t $DOCKER_IMAGE .

# Создание директории для выходного файла, если её нет
mkdir -p build

# Запуск контейнера Docker с подключением тома для выхода
docker run --rm -v "$(pwd)/build:/app/build" $DOCKER_IMAGE cp /app/merger /app/build/

# Проверка успешности копирования файла
if [ -f "build/merger" ]; then
    echo "Файл 'merger' успешно собран и помещен в директорию 'build'."
else
    echo "Ошибка: файл 'merger' не был скопирован в директорию 'build'."
    exit 1
fi