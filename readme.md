# go-wg-manager - менеджер Wireguard Server

---

## Для чего нужен

**gwg** - утилита командной строки для автоматического конфигурирования  и администрирования wireguard-сервера.
Поддерживает такие фунции как:

1) Автоматическая настройка конфигурации wireguard server;
2) Автоматическое изменение конфигурации сервера при добавлении пользователя;
3) Автоматическое изменение конфигурации сервера при удалении пользователя;
4) Автоматическое изменение конфигурации сервера при блокировке/разблокировке пользователя;
5) Просмотр состояния сервера через стандартную утилиту wg show;
6) Просмотр подробной статистики на основе стандартной утилиты wg show dump. (дорабатывается)

## Поддерживаемые платформы

- Любой дистрибутив linux на основе Debian.

## Установка

- Скачать архив с [желаемой](https://github.com/PavelMilanov/go-wg-manager/tags) версией:

```bash
wget https://github.com/PavelMilanov/go-wg-manager/releases/download/v0.2.4/gwg.tar
```

- Распаковать архив:

```bash
tar -xvf gwg.tar
```

- Запустить скрипт первичной настройки окружения gwg-manager и установки gwg
   ( **В конце установки будет предложено перезапустить сессию пользоватeля!** ):

```bash
./gwg-utils.sh install
```

## Обновление

- Запустить утилиту:

```bash
./gwg-utils.sh update v0.2.4
```

## Использование

- Просмотр состояния подключений:
![gwg show](./docs/images/show.png)

```bash
gwg show
```

- Просмотр подробной статистики:
![gwg stat](./docs/images/stat.png)

```bash
gwg stat
```

- Добавление пользователя:
![gwg add](./docs/images/add.png)

```bash
gwg add -name <alias>
```

- Удаление пользователя:
![gwg remove](./docs/images/remove.png)

```bash
gwg remove -name <alias>
```

- Блокировка пользователя:
![gwg block](docs/images/block.png)

```bash
gwg block -name <alias>
```

- Разблокировка пользователя:
![gwg unblock](./docs/images/unblock.png)

```bash
gwg unblock -name <alias>
```
