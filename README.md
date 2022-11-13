# User Balance Service

Микросервис для работы с балансом пользователей

Были выполнены все сценарии и доп. условия описанные в ТЗ:
* Работа с балансом (пополнение, снятие, перевод, получение)
* Работа с резервом денег (сам резерв, его отмена и подтверждение)
* Отчет сумм выручки по услугам в формате .csv 
* История баланса пользователя с пагинацией и сортировкой
* Тесты
* Swagger

## Запуск
Создаем .env файл, для удобства был создан пример, который можно скопировать
```
cp .env.example .env
```
Запускаем приложение
```
docker-compose up --build
```

## API

Более подробно API документацию можно посмотреть в Swagger по маршруту <b>/swagger</b>

![swagger](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/swagger.png)

Ниже будут приведены основные запросы в интерфейсе Postman-а:

* GET <b>/balance/</b>

Получение баланса пользователя

![balance_get](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/balance_get.png)

* POST <b>/balance/replenish/</b>

Пополнение баланса пользователя (создает новый баланс, если раньше не существовал)

![balance_replenish](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/balance_replenish.png)

* POST <b>/balance/reduce/</b>

Уменьшение баланса пользователя

![balance_reduce](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/balance_reduce.png)


* POST <b>/balance/transfer/</b>

Перевод денег с одного баланса на другой

![balance_transfer](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/balance_transfer.png)


* POST <b>/reservation/reserve/</b>

Резервирование денег на услугу

![reservation_reserve](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/reservation_reserve.png)

* POST <b>/reservation/cancel/</b>

Разрезервирование денег

![reservation_cancel](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/reservation_cancel.png)

* POST <b>/reservation/confirm/</b>

Подтверждение списывание денег за услугу

![reservation_confirm](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/reservation_confirm.png)


* GET <b>/history/</b>

История изменения баланса пользователя, есть необязательная пагинация (limit, offset), а также предусмотрена сортировка 
по сумме и дате (по умолчанию по дате в desc)

![history](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/history.png)


* POST <b>/report/</b>

Отчет суммарной выручки по услугам. Файл .csv пересоздается каждый раз только за текущий месяц

![report](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/report.png)

Пример скачанного отчета:

![report-example](https://github.com/garet2gis/user-balance-service/blob/master/documentation/images/csv.png)

## БД

[Файл со схемой данных](https://github.com/garet2gis/user-balance-service/blob/master/migrations/20221108113104_create_db_schema.up.sql)

Также стоит отметить, что все запросы с изменением баланса были выполнены в транзакциях
с уровнем изоляции Serializable

## Тесты

Для тестов создается отдельный изолированный контейнер, который заполняется
тестовыми данными. Внутри него прогоняются сами тесты и результат записывается
в файл cover.out

Такой вариант был выбран, чтобы более надежно протестировать БД, не используя моки

Команда для прогона тестов:
```
make test
```

Команда для более удобного просмотра покрытия:
```
make see-cover
```

Покрытие тестами составляет ≈ 70%

## Использованные технологии

1. Роутер [httprouter](https://github.com/julienschmidt/httprouter)
2. Postgres [pgx](https://github.com/jackc/pgx/v5)
3. Валидация [validator](http://github.com/go-playground/validator/v10)
4. Логгер [logrus](https://github.com/sirupsen/logrus)
5. Считывание конфига [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
6. Swagger [swag](https://github.com/swaggo/swag)

## Проблемы, с которыми столкнулся

В отчете по услугам нужно было выводить название услуги, которого не передавалось в запросе.

Данная проблема была решена созданием таблицы услуг (service), которая просто отображает ID услуги в название. Но 
для корректной работы перед запуском она должна быть заполнена, поэтому я так и сделал: заполнил ее фейковыми данными,
но думаю, что в production приложении стоило бы синхронизировать эту таблицу с сервисом услуг.
