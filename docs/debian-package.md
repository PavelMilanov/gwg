# Сборка и Debian-пакет

## Версия и Go

Рабочий каталог Go-модуля — `src`, точка входа — `src/cmd/gwu`.
Используйте Go toolchain из [go.mod](../src/go.mod).

Единственный источник версии —
[src/internal/gwucli/VERSION](../src/internal/gwucli/VERSION).
Формат: `vX.X.X`, без суффиксов и ведущих нулей. Версия встроена через
go:embed; Debian Version содержит те же числа без `v` и отдельной ревизии.
Не передавайте `version=...` в make.

Из корня репозитория:

```bash
make test
(cd src && go vet ./...)
make build
./gwu version
./gwu --help
```

`GO=/path/to/go` переопределяет компилятор для make.
`component=gwu` и `component=all` выбирают единственный компонент gwu.
По умолчанию собирается gwu; других бинарников нет.

## Debian

Нужны Go, make, debhelper (compat 13), dpkg-dev и fakeroot.
Проверьте build-зависимости из [control](../debian/gwu/control):
Make вызывает dpkg-buildpackage с `-d`, автоматическая проверка отключена.

```bash
make deb
```

Цель создаёт отдельный каталог `.build/deb/gwu/build.*/source`,
объединяет metadata из `debian/common` и `debian/gwu`, генерирует changelog
из VERSION и запускает сборку. Исходные файлы в debian не переписываются.
Прямой dpkg-buildpackage из корня не используется.

Результат по умолчанию: `vagrant/share/gwu_0.3.0_amd64.deb`.
Каталог `vagrant/share` служит только местом вывода, Vagrant не требуется.

```bash
make deb DEB_OUTPUT_DIR=dist
dpkg-deb --info dist/gwu_0.3.0_amd64.deb
dpkg-deb --contents dist/gwu_0.3.0_amd64.deb
```

Если используется переопределение GO, добавьте каталог этого компилятора в PATH
для Debian rules, которые вызывают go непосредственно.

Пакет содержит бинарник `/usr/bin/gwu`, man page, документацию,
лицензионные сведения и Bash completion. Установочных скриптов, меняющих сеть,
в пакете нет. Установка не запускает init и не устанавливает qos.service.
Runtime-зависимости перечислены в control; установка nftables требует
проверки поведения существующего firewall на целевой машине.

Пакеты gwg и gwu имеют разные имена и могут быть установлены одновременно.
Это не разрешает им одновременно управлять одним интерфейсом или TC.
Подробнее — [переход с gwg](migration.md).

## Release-архивы

```bash
make release
```

Создаёт Linux amd64 и 32-bit ARM архивы
`gwu_v0.3.0_linux_amd64.tar` и `gwu_v0.3.0_linux_arm.tar`.
Отдельные цели: `make amd`, `make arm`.
Архив содержит один бинарник gwu; распаковка сама не устанавливает зависимости.
`make clean` удаляет локальный бинарник gwu.
