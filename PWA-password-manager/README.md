###### СТУДЕНТ: Данг Фыонг Нам

###### ГРУППА: 932102

###### Эта работа вложена в Гитхаб

## lab2 Мобильное приложение

# REQUIRES:

## ➕ Структура:

    🟣 manifest.json, service-worker.js, icons ✅
    🟣 index.html (scipt: client.js)

## ➕ Install WPA:

    🟣 Button Install (will be hidden if app is installed, handled by beforeInstallPrompt Event) ✅
    🟣 Run offline with caches ✅

## ➕ Using storage:

    🟣 IndexedDB ✅

## ➕ Data and operations:

    🟣 Create Domain ✅
    🟣 Delete Domain (with all profiles inside domain) ✅
    🟣 Create Profile (with auto generated password) ✅
    🟣 Edit Profile Password ✅
    🟣 Show Info Profile Password (with old passwords before edited) ✅
    🟣 Delete Profile ✅

## ➕ Deadline:

    🟣 8 Jan 23:55 ✅

# RESULT:

## Запустить программу с коммандой "npx serve .", чтобы Service Worker смог работать, или просто открыть файл index.html на браузере (но будет без функции INSTALL)

## Приложение нужен интернет, потому что он использует "cdn tailwindcss" для стиля и "google font". По требованию это можно решать способом установить tailwindcss в папке проекта

### Обяснение файла scipt client.js: Он дает класс PasswordManagerApplication, который обеспечивает все методы, работающие с IndexedDB. Method displayList() отображает список домейнов и профилов на index.html строкой JSX. Каждый раз displayList() зовут, программа делает (document.removeEventListener) и заново (document.addEventListener), чтобы старые поведения не влияет на работу приложения.

### IndexedDB сохраняет 2 таблицы (ObjectStore): domains, profiles. Структура:

#### domains {id:number(auto increment), name:string (unique)}

#### profiles {id:number(auto increment), username:string, password:string, oldPassword[{password:string, savedAt:timestamp}], updatedAt:timestamp}
