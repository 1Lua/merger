#!/bin/bash

# Название директории для теста
TEST_DIR="test-git-repo"
BINARY="./../build/merger"

# Удаление старой тестовой директории, если она существует
echo "Удаление старой тестовой директории (если существует)..."
rm -rf $TEST_DIR

# Создание новой директории
echo "Создание новой директории: $TEST_DIR"
mkdir $TEST_DIR
cd $TEST_DIR

# Инициализация нового Git-репозитория
echo "Инициализация нового Git-репозитория в $TEST_DIR"
git init

# Создание начального коммита на master
echo "Создание начального коммита на ветке master"
echo "Initial commit" > file1.txt
git add file1.txt
git commit -m "Initial commit on master"
git tag "v1.0.0"

# Создание новой ветки dev и добавление коммита
echo "Создание новой ветки 'dev' и добавление коммита"
git checkout -b dev
echo "Commit on dev" >> file2.txt
git add file2.txt
git commit -m "Commit on dev"

# Возвращение на ветку master и добавление еще одного коммита
echo "Возвращение на ветку 'master' и добавление еще одного коммита"
git checkout master
echo "Second commit on master" >> file.txt
git add file.txt
git commit -m "Second commit on master"

echo "Граф коммитов:"
git log --graph --oneline --decorate

# Выполнение merge при помощи утилиты
echo "Выполнение слияния ветки 'dev' в 'master' с помощью утилиты"
$BINARY patch -s dev -t master

# Проверка результата
echo "Проверка результата слияния"
echo "Граф коммитов:"
git log --graph --oneline --decorate

# Проверка наличия тега
echo "Проверка наличия тега"
git tag