###### СТУДЕНТ: Данг Фыонг Нам
###### ГРУППА: 932102

###### Эта работа вложена в Гитхаб. Чтобы смотрет ответ на задания, переходите по гит-веткам (git branch)
##### Список гит-ветки:
###### main
###### lab1 (текущая ветка)


## REQUIRES:
  ### ➕ Using asyn + sync primitives:
    🟣 Using library asyncio (gather,run, sleep, async - await) ✅
  ### ➕ Download file html:
    🟣 Download with URL (provide in command line as an arguement) ✅
    🟣 Save to current folder with original name. ✅
  ### ➕ count taken bytes per second:
    🟣 Every second gives output of size of taken data. ✅
  ### ➕ Base language:
    🟣 Using Python ✅
  ### ➕ Deadline:
    🟣 8 Jan 23:55 ✅

## RECOMMENDATION:
  ### 🟢 Using http.client (HTTPConnection,request, getResponse, close)✅

## RESULT:
### Запустить программу с коммандой "python dang.py"
### Menu появится и предложит 2 варианта использования:
#### 1) Ввести URL
##### Если введенный URL правильный, то программа начнется скачание файл и сохранить с орининальным именем (title). При процессе скачания выводит через каждую секунду в концоле сообщение, сколько битов принятых в конкретное время. После процесса скачания программа вернется в MENU.
###### URL правильный: response о сервера - 200 ОК, content-type: text/html. Правильная формата: "http" или "https" + "://" + domain + "/" +path
###### Процесс скачание: (WHILE LOOP до того, когда будет буффер = 0, значит до конца response) читать response в буффер (размер 2MB) и записать буффер в файл (tempfile). Когда WHILE LOOP завершен, то есть tempfile записан, начнет поиск в нем title, переименовать файл полученным ответом от поиска.
###### Поиск title: regex(<title>(.?)</title>), полученый ответ будет проверен, чтобы удалить от него особые символы
#### 2) Выйти - Ввести "q"
