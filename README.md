# Финальная задача по курсу Golang от Яндекс лицея
Краткое тз:

Пользователь хочет считать арифметические выражения. Он вводит строку 2 + 2 * 2 и хочет получить в ответ 6. Но наши операции сложения и умножения (также деления и вычитания) выполняются "очень-очень" долго. Поэтому вариант, при котором пользователь делает http-запрос и получает в качетсве ответа результат, невозможна. Более того: вычисление каждой такой операции в нашей "альтернативной реальности" занимает "гигантские" вычислительные мощности. Соответственно, каждое действие мы должны уметь выполнять отдельно и масштабировать эту систему можем добавлением вычислительных мощностей в нашу систему в виде новых "машин"

# Описание принципа работы
Мое решение содержит несколько независимых частей, которые могут быть запущены отдельно: 
 - фронтенд
 - базу данных
 - оркестратор
 - вычислительный сервер

Все они поднимаются при помощи docker-compose. Фронтенд отвечает за запуск веб-страницы, в ней работает java-script, отправляющий запросы на оркестратор. Оркестратор является центральным узлом, который записывает задачи в базу данных, сохраняет настроки, выдает задачи вычислительному серверу и обрабатывает ошибки. Вычислительный сервер запускает указанное количество вычислителей, каждый из вычислителей работает в отдельной горутине, а так же каждую задачу, которую принимает, разбивает на подзадачи, которые он может решить параллельно.

## Фронтэенд
Я не силен во фронте, по этому сделал достаточно простой сайт имеющий четыре вкладки:
 - Поле для введения выражения, а так жее окошко с ответом от сервера
 - Вкладка со списком задач, их статусом и датой завершения вычисления
 - Поля для ввода времени выполнения операций
 - Список с зарешистрированными вычислителями, их статусами и выражениями, которые они считают

## Оркестратор
Оркестратор представляет собой API с различными эндпоинтами вот их список:
 - 123
 - 456

Оркестратор при запуске создает подключение к базе данных, и если нужно, то создает в ней необходимые таблицы. Затем если загружает настройи из базы данных, и запускает исполнителей, каждый из которых отвечает за свой эндпоинт.

Оркестратор сам не отправляет запросов, любой кто хочет получить данные о работе системы или отправить задачу должен отправить HTTP запрос на откестратор. Оркестратор в качестве способа обмена данными использует только JSON в теле запроса и в теле ответа. 

Получая задачу, откестратор кладет ее в таблицу базы данных. Когда вычислитель просит задачу, оркестратор меняет статус задачи в базе, после чего выдает ее вычислителю, при этом запоминая, какой вычислитель какую хадачу взял. Когла вычислитель делает запрос с ответом, оркестратор меняет статус задачи в базе данных и записывает ответ.

## Вычислительный сервер
С ним я несколько оподливился, так как парсер, который я написал самостоятельно, малофункциональный и не самый оптимальный. (Простите. Я старался)
Вычислительный сервер запускает указанное количество вычислителей. Из за парсера вычислитель умеет только считать выражения из целых чисел без скобок. Поддерживается только сложение, вычитание, деление и умножение.

Принцип деления выражения на подзадачи: 
Возьмем выражение ```1*2+2-3*4/7-9```, в нем имеют приоритет операции ```1*2``` и ```3*4/7```, то есть группы в котрых только умножения и деления, такие группы будут запущены в отдельных горутинах, когда все группы будут подсчитаны, можно выпонять операции второго приоритета, то есть начнется вычисление выражения ```2+2-1.714285-9```. Если умножение выполняется за 10 секунд, деление за 2 секунды, сложение за 5, а вычитание за 1, то такое выражение будет подсчитано за ```12+5+1+1=19``` секунды,  так как ```1*2``` и ```3*4/7``` считаются параллельно, но ```3*4/7``` считается на две секунды дольше, итого 12.

# Запуск и тестирование
Все упаковано в docker-compose. Для запуска в Linux нужжно ввести команду:
```
docker-compose up -d
```
Если интересно посмотреть логи, то:
```
docker-compose up -d
```
Для остановки:
```
docker-compose down
```
Важно: команда ```docker-compose down``` удалит контейнеры, то есть данные в базе данных потеряются. Для остановки без удаления используйте запуск с логами с последующим отклбчением через Ctrl+C в терминале.

Установить докер:
```
curl -sSL https://get.docker.com | sh
```
```
sudo usermod -aG docker $(whoami)
```

Для просмотра веб страуицы переидите по: (http://localhost:8081/frontendSite)
