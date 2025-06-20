# Worker Pool

Реализация пула воркеров на Go с возможностью динамического добавления и удаления воркеров во время выполнения.

## Описание

Программа имитирует обработку задач со строками, поступающими из текстового файла. Воркеры читают задачи из общего канала и обрабатывают их. В качестве примера, используется простой текстовый файл с оглавлением популярной статьи.

## Установка и использование

1. **Клонировать репозиторий:**

    ```sh
    git clone https://github.com/senyabanana/go-workerpool.git
    cd go-workerpool
    ```
   
2. **Запустить программу:**

    ```sh
    go run cmd/main.go
    ```
   или же

    ```sh
    go build cmd/main.go
    ./main
    ```
   
3. **Подход:**

- Задачи подаются с интервалом 500мс
- Воркеры добавляются или удаляются каждые 3 секунды случайным образом
- Программа завершает работу через 50 секунд или после подачи всех задач
- Пустые строки из файла игнорируются

## Структура

```
.
├── cmd
│   └── main.go                  # точка входа
├── internal
│   ├── fileinput                # пакет, реализующий чтение из файла
│   │   └── reader.go
│   ├── worker                   # реализация воркера и его запуска
│   │   └── worker.go
│   └── workerpool               # реализация пула воркеров и его методов
│       └── pool.go
├── 50_shades_of_go_example.txt  # пример текстового файла
├── go.mod
└── README.md
```