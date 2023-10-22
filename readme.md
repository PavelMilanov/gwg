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
6) Просмотр подробной статистики на основе стандартной утилиты wg show dump.

## Поддерживаемые платформы

- Любой дистрибутив linux на основе Debian.

## Установка

- Скачать архив с [желаемой](https://github.com/PavelMilanov/gwg/releases) версией и поддерживаемой архитектурой:

```bash
wget https://github.com/PavelMilanov/gwg/releases/download/v0.2.5/gwg.linux_amd64.tar
```

- Распаковать архив:

```bash
tar -xvf gwg-linux_amd64.tar
```

- Перенести исполняемый файл в директорию /usr/bin:

```bash
sudo cp gwg /usr/bin/
```

- Запустить установку:

```bash
./gwg init
```

## Обновление

- В разработке

## Использование

- Просмотр общего функионала:

```bash
gwg -h
```

![gwg help](./docs/images/menu.png)

- Просмотр состояния подключений:

```bash
gwg show
```

![gwg show](./docs/images/show.png)

- Просмотр подробной статистики:

```bash
gwg stat
```

![gwg stat](./docs/images/stat.png)

- Добавление пользователя:

```bash
gwg add -name <alias>
```

![gwg add](./docs/images/add.png)

- Удаление пользователя:

```bash
gwg remove -name <alias>
```

![gwg remove](./docs/images/remove.png)

- Блокировка пользователя:

```bash
gwg block -name <alias>
```

![gwg block](docs/images/block.png)

- Разблокировка пользователя:

```bash
gwg unblock -name <alias>
```

![gwg unblock](./docs/images/unblock.png)
