# Debian-пакет gwg

## Разделение ответственности

Debian-пакет выполняет только две задачи:

1. Устанавливает исполняемый файл `/usr/bin/gwg` и документацию.
2. Объявляет runtime-зависимости, которые APT устанавливает автоматически.

Пакет не создает WireGuard-интерфейс, не генерирует ключи, не изменяет
`sysctl`, не запускает службы и не вызывает `gwg init` из maintainer scripts.

Инициализация сервера выполняется пользователем отдельно:

```bash
sudo gwg init
```

Команда создает закрытые рабочие каталоги в `/etc/wireguard`, записывает
`/etc/sysctl.d/90-gwg.conf`, включает IPv4 forwarding и устанавливает сервер
`wg0` с сетью `10.0.0.1/24` и UDP-портом `51830`.

## Runtime-зависимости

Зависимости перечислены в `debian/control`:

| Пакет | Назначение |
| --- | --- |
| `wireguard-tools` | Команды `wg` и `wg-quick` |
| `iproute2` | Команды `ip` и `tc` |
| `iptables` | NAT и маршрутизация WireGuard |
| `procps` | Команда `sysctl` |
| `systemd` | Управление `wg-quick` и TC-службой |
| `sudo` | Выполнение системных команд текущей реализацией gwg |

Зависимости скачиваются только при установке пакета через APT. Программа
`gwg` не запускает `apt` или `dpkg`.

## Подготовка сборочной системы

Для сборки нужны `debhelper`, `dpkg-dev`, `fakeroot` и Go 1.26 или новее, как
указано в `src/go.mod` и `debian/control`:

```bash
sudo apt update
sudo apt install -y build-essential debhelper dpkg-dev fakeroot golang-go
go version
```

Если версия Go в используемом выпуске Debian ниже требуемой, необходимо
установить подходящий Go toolchain отдельно. Поскольку такой toolchain не
зарегистрирован в базе `dpkg`, после установки остальных build-зависимостей
проверка выполняется с отключенной проверкой `Build-Depends`:

```bash
dpkg-buildpackage -us -uc -b -d
```

Опция `-d` допустима только после ручной проверки `go version` и установки
остальных build-зависимостей. В окружении, где доступен Debian-пакет
`golang-any` версии 1.26 или новее, следует использовать обычную команду без
`-d`.

## Версия пакета

Версия задается первой записью `debian/changelog`:

```text
gwg (0.2.6.4-1) unstable; urgency=medium
```

Здесь `0.2.6.4` — версия программы, а `-1` — ревизия Debian-пакета. Версия
программы автоматически встраивается в бинарник через `main.VERSION`.

Для нового выпуска следует добавить запись командой:

```bash
dch -v 0.2.6.5-1 "Release 0.2.6.5"
```

## Сборка

Из корня репозитория:

```bash
dpkg-buildpackage -us -uc -b
```

Перед созданием пакета `debian/rules` переходит в `src` и выполняет:

```bash
cd src
go test ./...
go build -trimpath -buildmode=pie ...
```

Готовые файлы создаются в родительском каталоге, например:

```text
../gwg_0.2.6.4-1_amd64.deb
../gwg-dbgsym_0.2.6.4-1_amd64.deb
../gwg_0.2.6.4-1_amd64.changes
```

## Проверка содержимого

```bash
dpkg-deb --info ../gwg_0.2.6.4-1_amd64.deb
dpkg-deb --contents ../gwg_0.2.6.4-1_amd64.deb
lintian ../gwg_0.2.6.4-1_amd64.changes
```

В основном пакете должны находиться `/usr/bin/gwg`, man-страница и
документация. В пакете не должно быть `postinst`, `preinst`, `prerm` или
`postrm`, выполняющих настройку сервера.

## Установка

Для локального `.deb` следует использовать APT, а не `dpkg -i`, поскольку APT
разрешает и скачивает зависимости:

```bash
sudo apt install ./gwg_0.2.6.4-1_amd64.deb
```

После установки можно проверить зависимости и версию:

```bash
dpkg-query -W gwg wireguard-tools iproute2 iptables procps systemd sudo
gwg version
```

Сервер на этом этапе еще не настроен.

## Инициализация сервера

Для стандартной конфигурации:

```bash
sudo gwg init
```

Для явного задания параметров вместо `init`:

```bash
sudo gwg install -name wg0 -network 10.0.0.1/24 -port 51830
```

После инициализации:

```bash
sudo systemctl status wg-quick@wg0.service
sudo wg show wg0
```

## Обновление и удаление

Обновление заменяет бинарник, но не запускает инициализацию и не изменяет
файлы в `/etc/wireguard`:

```bash
sudo apt install ./gwg_NEW_VERSION_amd64.deb
```

Удаление пакета:

```bash
sudo apt remove gwg
```

Поскольку конфигурации создаются самой программой, а не пакетом, APT их не
удаляет. Удалять WireGuard-конфигурацию и ключи следует только отдельной
осознанной административной операцией.

## Проверка в Vagrant

Скопируйте пакет в общий каталог и установите его на серверной VM:

```bash
cp ../gwg_0.2.6.4-1_amd64.deb vagrant/share/
vagrant up server
vagrant ssh server
sudo apt install /home/vagrant/share/gwg_0.2.6.4-1_amd64.deb
sudo gwg init
```

При проверке установки VM должна иметь доступ к Debian-репозиториям для
скачивания объявленных зависимостей.
