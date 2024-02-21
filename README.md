# Распределенный вычислитель арифметических выражений
* Автор - https://t.me/K_alashni_k (если что-то не работает или не запускается - напиши мне пж)
* Постараюсь всё максимально объяснить, но если что-то не понятно, то пишите в тг.
## Установка
1. Скачать этот репозиторий на своё устройство, открыть в вашей **IDE** (желательно GoLand)
2. Скачать **Docker desktop v4.27.2 (Если уже установлен то проверьте версию, в 4.27.1 есть баг из за которого проект не запускается)** https://www.docker.com/products/docker-desktop/
3. Запустить установку **Docker desktop**
4. Запустить **Docker desktop**, подождать пока он всё сделает (не уверен обязательно ли там заходить в аккаунт). **Он должен оставаться в фоне во время работы программы.**
5. В терминале открыть корневую папку с кодом (`cd <путь>`)
6. Прописать `docker`, если нету такой команды то вот:
<details><summary>Решение с хабра</summary>
Проверьте переменные зависимости. В переменной PATH мог не прописаться путь до docker.exe. Найдите путь до docker.exe (обычно в папке bin) и добавьте путь в переменную PATH
</details>

## Запуск
1. Запустить с помощью: `docker-compose up -d`
2. Подождите (иногда долго, возможно даже 10 минут)
3. По окончанию загрузки должен работать http://localhost/ (http://localhost:80/). Когда на странице http://localhost/agents появятся агенты значит всё загрузилось.
4. После docker-compose up -d можете перезапустить или выключить один из микросервисов. 
   * `docker-compose stop <название микросервиса / если не писать то все>`
   * `docker-compose restart <название микросервиса / если не писать то все>`
   * `docker-compose up <название микросервиса / если не писать то все> -d`
### Имеющиеся микросервисы:
* **orchestrator** - оркестратор и API
* **user_server** - сервер с FrontEnd, просто возвращает html
* **agents** - запускает агентов, количество задаётся в 55 строке [docker-compose.yaml](docker-compose.yml) (там устанавливается environment)
* **postgres** - база данных
* **rabbitmq** - брокер сообщений
* **pgadmin** - Админ панель базы данных, доступна на http://localhost:5050/login (для входа `admin@admin.com`, `root`)

## Использование
1. Инструкции для API не нужны ведь есть интерфейс на http://localhost:80/, API находится на http://localhost:8080/ , инструкции к нему есть в схеме
2. На ввод может быть любое правильно математическое выражение, оно может содержать целые числа или с плавающей точкой, скобки, 4 оператора: +, -, *, /. Перед числами ставить знаки нельзя, только между ними (`-1 + 2` нельзя, можно только `(0 - 1) + 2`)
## Схема работы всего этого
https://excalidraw.com/#json=RjMvgDnlYwJ8BOajzF8cV,C9jkJjFo4I3VkJaYcgiXcw

![image](https://i.imgur.com/2zd4tXx.png)

## А теперь по критериям
0. Необходимые требования:
- Существует Readme документ, в котором описано, как запустить систему и как ей пользоваться.
-   Это может быть docker-compose, makefile, подробная инструкция - на ваш вкус
  - Если вы предоставляете только http-api, то
    - в Readme описаны примеры запросов с помощью curl-a или любым дргуми понятными образом
    - примеры полны и понятно как их запустить

  Этот пункт дает 10 баллов. Без наличия такого файла - решение не проверяется.

**✓ Есть**

1. Программа запускается и все примеры с вычислением арифметических выражений корректно работают - 10 баллов

**✓ Работают**


2. Программа запускается и выполняются произвольные примеры с вычислением арифметических выражений - 10 баллов

**✓ Да** 

3. Можно перезапустить любой компонент системы и система корректно обработает перезапуск (результаты сохранены, система продолжает работать) - 10 баллов

**? На решение проверяющего.** Оркестратор и сервер с FrontEnd перезапустятся без проблем, всё хранится в **PostgresSQL**. При перезапуске агента он завершает своё выражение и выводит для него ошибку. Перезапускать **PostegreSQL** и **RabbitMQ** нельзя.

4. Система предосталяет графический интерфейс для вычисления арифметических выражений - 10 баллов

**✓ Да, даже **Bootstrap'а** для стиля добавил**

5. Реализован мониторинг воркеров - 20 баллов

**✓ Да, API http://localhost:8080/agents возвращает статусы всех агентов и что они выполняют**

6. Реализован интерфейс для мориторинга воркеров - 10 баллов

**✓ Да http://localhost/agents**

7. Вам понятна кодовая база и структура проекта - 10 баллов (это субъективный критерий, но чем проще ваше решение - тем лучше).
Проверяющий в этом пункте честно отвечает на вопрос: "Смогу я сделать пулл-реквест в проект без нервного срыва"

**? На решение проверяющего.**

8. У системы есть документация со схемами, которая наглядно отвечает на вопрос: "Как это все работает" - 10 баллов

**✓ Да, https://excalidraw.com/#json=RjMvgDnlYwJ8BOajzF8cV,C9jkJjFo4I3VkJaYcgiXcw (или в корне проекта или выше на imgur)**

9. Выражение должно иметь возможность выполняться разными агентами - 10 баллов
   Итого 110 баллов

**✓ Да, но...** к примеру:

* `(1 + 2) * (4 + 5 * 6)`  -  Здесь `(1 + 2)` и `5 * 6` будут выполнятся на двух разных агентах
* `1 * 5 + 4 * 2 - 1`      -  Здесь `1 * 5` и `4 * 2` выполнятся параллельно
* Но `1 + 2 + 3 + 4 + 5 + 6` - Здесь из-за особенности алгоритма перевода в постфиксную форму оно преобразуется в последовательность операций `(1 + (2 + (3 + (4 + (5 + 6)))))`.
* В общем: операции могут выполняться параллельно, только если они на разном уровне приоритета, например `1 + 1 + 1 + 1` не будет выполняться параллельно, как и `(5 * 4 * 3 * 2)`
* Но по факту возможность имеет. **✓**
