# Makefile

# Название исполняемого файла
BINARY=merger

# Go-файлы проекта
SOURCES=main.go gitUtility.go

# Директория для установки
INSTALL_DIR=/usr/local/bin

# Команда сборки
build:
	go build -o ./build/$(BINARY) $(SOURCES)

# Команда запуска
run: build
	./build$(BINARY) -s dev -t master --tag v1.0.0

# Команда установки
install: build
	install -m 0755 ./build/$(BINARY) $(INSTALL_DIR)

# Команда очистки
clean:
	rm -f $(BINARY)

# Псевдонимы
.PHONY: build run install clean
