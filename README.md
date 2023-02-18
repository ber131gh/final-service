### Сервис собирающий данные о состоянии различных систем компании для дальнейшего возрата структурированного ответа, содержащего в себе StatusPage, географию и статусы систем.

#### Подготовка сервиса

1. Создайте директорию final_service в директории src вашего окружения golang
2. Склонируйте этот репозиторий в директорию final_service с помощью системы контроля версий git (например, перейдя в директрию выполните в консоли команду git clone git@github.com:ber131gh/final-service.git


### Запуск и проверка результата работы дипломного проекта

1. Разделите терминал среды разработки, который вы используете, на 2 части. 
2. Сервис содержит в себе проект simulator, через файлы и API которого, приложение получает доступ к данным различных систем.
3. Сперва необходимо запустить simulator. В одной части терминала перейдите в папку `\final_exam\simulator\skillbox-diploma>`, затем в ней выполните команду `go run main.go`.
4. Затем запустим сервис. Для этого в другой части терминала перейдите в корневую папку сервиса `\Service`, в которую был склонирован проект. Выполните в ней команду `go run main.go`.
5. Откройте в браузере `http://localhost:8282/systemsstatus`, при запущенном приложении и симуляторе. На странице браузера должна отобразиться информация о статусе обработки данных систем, географии и состоянии различных систем.


#### Особенности работы приложения

Приложение запускает сервер и слушает соединение: `localhost:8282`.
К серверу прикреплен роутер, к которому добавлено 2 обработчика:
 `/` обрабатывается функция handleConnection, целью которой является первичное тестирование обработки запросов, возвращает только слово `OK`.
 `/systemsstatus` обрабатывается функция getSystemsData, возвращающая конечную структуру с отфильтрованными данными в формате json.

При запросе по адресу `http://localhost:8282/systemsstatus`, приложение находит и считывает данные одних систем из файлов симулятора, других систем через API симулятора.

Данные систем получаемые через API
* http://127.0.0.1:8383/mms - данные по системе MMS
* http://127.0.0.1:8383/support - данные по системе Support
* http://127.0.0.1:8383/accendent - данные по системе инцидентов
* http://127.0.0.1:8383/test - заглушка для первичной демонстрации StatusPage синтетическими данными

Данные систем получаемые через файлы
* \service\simulator\skillbox-diploma\sms.data
* \service\simulator\skillbox-diploma\voice.data
* \service\simulator\skillbox-diploma\email.data
* \service\simulator\skillbox-diploma\billing.data

Завершения работы приложения(Graceful Shutdown): приложение ожидает сигнал Interrupt(сочетание клавиш `ctrl+C`), после чего закрывает сервер.

При каждом запуске симулятора, он генерирует новые данные для того, чтобы можно было произвести отладку приложения на разных данных.
Информацию по работе simulator можно найти в директории проекта: `\Service\simulator\skillbox-diploma\README.md`.