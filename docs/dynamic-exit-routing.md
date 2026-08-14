# Динамическая маршрутизация через несколько внешних шлюзов

## Назначение

В этой схеме центральный WireGuard-сервер обслуживает обычных VPN-клиентов и
использует два независимых внешних узла как интернет-шлюзы:

```text
                              wg-exit1                  eth0
VPN clients -> wg0 -> server ------------> external-1 ------> Internet
                    |
                    |         wg-exit2                  eth0
                    +--------------------> external-2 ------> Internet
```

FRR/BGP поддерживает одновременно несколько равноправных default route. Linux
распределяет новые соединения между доступными внешними шлюзами. При потере
BGP-сессии шлюз автоматически исключается из ECMP-группы, а после восстановления
снова включается в нее.

## Почему нужны отдельные интерфейсы

`AllowedIPs` в WireGuard одновременно определяет допустимые source-адреса peer
и выбор peer для исходящего пакета. Два peer с `AllowedIPs = 0.0.0.0/0` на одном
интерфейсе не позволяют FRR выбирать следующий шлюз.

Поэтому используются три интерфейса:

| Узел | Интерфейс | Адрес | Назначение |
| --- | --- | --- | --- |
| server | `wg0` | `10.0.0.1/24` | Обычные VPN-клиенты |
| server | `wg-exit1` | `10.255.1.1/30` | Туннель до external-1 |
| external-1 | `wg-exit1` | `10.255.1.2/30` | Первый интернет-шлюз |
| server | `wg-exit2` | `10.255.2.1/30` | Туннель до external-2 |
| external-2 | `wg-exit2` | `10.255.2.2/30` | Второй интернет-шлюз |

Интерфейсы `wg-exit1` и `wg-exit2` используют разные ключи и UDP-порты.
Внешние шлюзы необходимо удалить из списка peers интерфейса `wg0`.

## Модель балансировки

Балансировка должна выполняться по соединениям, а не по отдельным пакетам.
Пакеты одного TCP- или UDP-потока должны проходить через один внешний шлюз,
поскольку каждый шлюз выполняет NAT в собственный публичный адрес. Попакетная
балансировка приведет к переупорядочиванию пакетов и смене исходного публичного
адреса внутри соединения.

На server используется ECMP с хешированием по 5-tuple: source IP, destination
IP, протокол, source port и destination port. Это дает следующие свойства:

- новые соединения распределяются между всеми доступными шлюзами;
- пакеты одного соединения выбирают один и тот же шлюз;
- BGP/BFD исключает недоступный шлюз из набора;
- возвращенный в работу шлюз начинает получать новые соединения.

ECMP выравнивает количество потоков, но не измеряет фактическую загрузку каналов.
Шлюзы с разной пропускной способностью требуют весов. Балансировка по текущей
загрузке требует отдельного контроллера и описана ниже.

## Подготовка ключей

На server:

```bash
sudo -i
umask 077
wg genkey | tee /etc/wireguard/wg-exit1.key | wg pubkey > /etc/wireguard/wg-exit1.pub
wg genkey | tee /etc/wireguard/wg-exit2.key | wg pubkey > /etc/wireguard/wg-exit2.pub
exit
```

На external-1:

```bash
sudo -i
umask 077
wg genkey | tee /etc/wireguard/wg-exit1.key | wg pubkey > /etc/wireguard/wg-exit1.pub
exit
```

На external-2 аналогично создаются `wg-exit2.key` и `wg-exit2.pub`. Между
узлами передаются только открытые ключи `.pub`.

## WireGuard на центральном сервере

Файл `/etc/wireguard/wg-exit1.conf`:

```ini
[Interface]
PrivateKey = SERVER_EXIT1_PRIVATE_KEY
Address = 10.255.1.1/30
ListenPort = 51831
Table = off
PostUp = iptables -A FORWARD -i %i -o wg0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
PostDown = iptables -D FORWARD -i %i -o wg0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

[Peer]
PublicKey = EXTERNAL1_PUBLIC_KEY
AllowedIPs = 0.0.0.0/0
```

Файл `/etc/wireguard/wg-exit2.conf`:

```ini
[Interface]
PrivateKey = SERVER_EXIT2_PRIVATE_KEY
Address = 10.255.2.1/30
ListenPort = 51832
Table = off
PostUp = iptables -A FORWARD -i %i -o wg0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
PostDown = iptables -D FORWARD -i %i -o wg0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

[Peer]
PublicKey = EXTERNAL2_PUBLIC_KEY
AllowedIPs = 0.0.0.0/0
```

`Table = off` запрещает `wg-quick` создавать маршруты из `AllowedIPs`.
Маршруты устанавливает FRR. На firewall центрального сервера должны быть
разрешены UDP-порты `51831` и `51832`.

## Конфигурация external-1

Файл `/etc/wireguard/wg-exit1.conf`:

```ini
[Interface]
PrivateKey = EXTERNAL1_PRIVATE_KEY
Address = 10.255.1.2/30
Table = off
PostUp = iptables -A FORWARD -i %i -o eth0 -j ACCEPT; iptables -A FORWARD -i eth0 -o %i -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT; iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -o eth0 -j ACCEPT; iptables -D FORWARD -i eth0 -o %i -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT; iptables -t nat -D POSTROUTING -s 10.0.0.0/24 -o eth0 -j MASQUERADE

[Peer]
PublicKey = SERVER_EXIT1_PUBLIC_KEY
Endpoint = SERVER_PUBLIC_IP:51831
AllowedIPs = 10.255.1.1/32, 10.0.0.0/24
PersistentKeepalive = 25
```

Если внешний интерфейс называется не `eth0`, его необходимо заменить в обеих
строках hooks.

## Конфигурация external-2

Файл `/etc/wireguard/wg-exit2.conf` отличается адресами, портом и ключами:

```ini
[Interface]
PrivateKey = EXTERNAL2_PRIVATE_KEY
Address = 10.255.2.2/30
Table = off
PostUp = iptables -A FORWARD -i %i -o eth0 -j ACCEPT; iptables -A FORWARD -i eth0 -o %i -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT; iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -o eth0 -j ACCEPT; iptables -D FORWARD -i eth0 -o %i -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT; iptables -t nat -D POSTROUTING -s 10.0.0.0/24 -o eth0 -j MASQUERADE

[Peer]
PublicKey = SERVER_EXIT2_PUBLIC_KEY
Endpoint = SERVER_PUBLIC_IP:51832
AllowedIPs = 10.255.2.1/32, 10.0.0.0/24
PersistentKeepalive = 25
```

## IPv4 forwarding

На external-1 и external-2:

```bash
echo 'net.ipv4.ip_forward=1' | sudo tee /etc/sysctl.d/90-gwg-forward.conf
sudo sysctl --system
```

На server дополнительно включается L4-хеширование ECMP:

```bash
sudo tee /etc/sysctl.d/90-gwg-forward.conf >/dev/null <<'EOF'
net.ipv4.ip_forward=1
net.ipv4.fib_multipath_hash_policy=1
EOF
sudo sysctl --system
```

Значение `1` для `fib_multipath_hash_policy` выбирает 5-tuple. При значении `0`
Linux учитывает только L3-поля, поэтому один VPN-клиент может получать менее
равномерное распределение соединений.

Проверка:

```bash
sysctl net.ipv4.ip_forward
```

## Запуск туннелей

На server:

```bash
sudo systemctl enable --now wg-quick@wg-exit1
sudo systemctl enable --now wg-quick@wg-exit2
```

На external-1 и external-2 запускается соответствующий интерфейс:

```bash
sudo systemctl enable --now wg-quick@wg-exit1
sudo systemctl enable --now wg-quick@wg-exit2
```

На server должны работать проверки:

```bash
sudo wg show wg-exit1
sudo wg show wg-exit2
ping -c 3 10.255.1.2
ping -c 3 10.255.2.2
ip route show 10.255.1.0/30
ip route show 10.255.2.0/30
```

До запуска FRR default route через эти интерфейсы появляться не должен.

## Защита основного маршрута server

FRR установит полученный BGP default route в основную таблицу. Чтобы ответы с
физического uplink server не ушли обратно через exit-туннель, исходный маршрут
необходимо сохранить в отдельной таблице. Это особенно важно, если публичный
адрес предоставляется через NAT, а локально назначен частный source-адрес.

Сначала определить параметры текущего default route:

```bash
ip route show default
ip route get 1.1.1.1
```

Добавить таблицу и маршрут, заменив значения примера фактическими:

```bash
grep -qE '^200[[:space:]]+uplink$' /etc/iproute2/rt_tables || \
  echo '200 uplink' | sudo tee -a /etc/iproute2/rt_tables
sudo ip route replace table uplink default via PHYSICAL_GATEWAY dev PHYSICAL_INTERFACE
sudo ip rule add priority 100 from SERVER_UPLINK_SOURCE_IP/32 lookup uplink
```

Правило должно создаваться вместе с интерфейсом сервера и удаляться при его
остановке. Использование `table main` для этого правила недостаточно: после
запуска BGP именно в `main` появится новый default route.

Проверка:

```bash
ip rule show
ip route show table uplink
ip route get 1.1.1.1 from SERVER_UPLINK_SOURCE_IP
```

Последняя команда должна показывать физический интерфейс, а не `wg-exit1` или
`wg-exit2`.

## Установка FRR

На всех трех узлах:

```bash
sudo apt update
sudo apt install -y frr
sudo sed -i 's/^bgpd=no/bgpd=yes/' /etc/frr/daemons
sudo sed -i 's/^bfdd=no/bfdd=yes/' /etc/frr/daemons
sudo systemctl restart frr
```

Используемые AS:

| Узел | AS |
| --- | --- |
| server | `65000` |
| external-1 | `65101` |
| external-2 | `65102` |

## FRR на server: active/active

Основные элементы `/etc/frr/frr.conf`:

```text
frr defaults traditional
hostname server
service integrated-vtysh-config
!
ip prefix-list DEFAULT-ONLY seq 10 permit 0.0.0.0/0
ip prefix-list VPN-ONLY seq 10 permit 10.0.0.0/24
!
route-map EXIT-IN permit 10
 match ip address prefix-list DEFAULT-ONLY
!
route-map VPN-OUT permit 10
 match ip address prefix-list VPN-ONLY
!
router bgp 65000
 bgp router-id 10.0.0.1
 bgp bestpath as-path multipath-relax
 neighbor 10.255.1.2 remote-as 65101
 neighbor 10.255.1.2 bfd
 neighbor 10.255.2.2 remote-as 65102
 neighbor 10.255.2.2 bfd
 !
 address-family ipv4 unicast
  network 10.0.0.0/24
  maximum-paths 16
  neighbor 10.255.1.2 activate
  neighbor 10.255.1.2 route-map EXIT-IN in
  neighbor 10.255.1.2 route-map VPN-OUT out
  neighbor 10.255.2.2 activate
  neighbor 10.255.2.2 route-map EXIT-IN in
  neighbor 10.255.2.2 route-map VPN-OUT out
 exit-address-family
!
```

Префикс-листы не позволяют внешним узлам передать серверу произвольные
маршруты. Оба default route имеют одинаковые BGP-атрибуты и входят в ECMP.
`multipath-relax` разрешает объединить пути из разных внешних AS, а
`maximum-paths 16` ограничивает группу шестнадцатью шлюзами.

## FRR на external-1

```text
frr defaults traditional
hostname external-1
service integrated-vtysh-config
!
ip prefix-list DEFAULT-ONLY seq 10 permit 0.0.0.0/0
ip prefix-list VPN-ONLY seq 10 permit 10.0.0.0/24
!
route-map DEFAULT-OUT permit 10
 match ip address prefix-list DEFAULT-ONLY
!
route-map VPN-IN permit 10
 match ip address prefix-list VPN-ONLY
!
router bgp 65101
 bgp router-id 10.255.1.2
 neighbor 10.255.1.1 remote-as 65000
 neighbor 10.255.1.1 bfd
 !
 address-family ipv4 unicast
  network 0.0.0.0/0
  neighbor 10.255.1.1 activate
  neighbor 10.255.1.1 route-map VPN-IN in
  neighbor 10.255.1.1 route-map DEFAULT-OUT out
 exit-address-family
!
```

Default route анонсируется, пока он присутствует в локальной таблице
external-1 и BGP-сессия работает.

## FRR на external-2

Конфигурация аналогична external-1 со следующими заменами:

```text
hostname external-2
router bgp 65102
 bgp router-id 10.255.2.2
 neighbor 10.255.2.1 remote-as 65000
 neighbor 10.255.2.1 bfd
```

Внутри `address-family ipv4 unicast` все обращения к `10.255.1.1` заменяются
на `10.255.2.1`.

После изменения конфигурации на каждом узле:

```bash
sudo chown frr:frr /etc/frr/frr.conf
sudo chmod 0640 /etc/frr/frr.conf
sudo systemctl restart frr
```

## Проверка BGP и переключения

На server:

```bash
sudo vtysh -c 'show bgp ipv4 unicast summary'
sudo vtysh -c 'show bgp ipv4 unicast 0.0.0.0/0'
sudo vtysh -c 'show bfd peers'
ip route show default
ip route get 8.8.8.8 from 10.0.0.2
ip route get 8.8.8.8 from 10.0.0.2 ipproto tcp sport 40001 dport 443
ip route get 8.8.8.8 from 10.0.0.2 ipproto tcp sport 40002 dport 443
```

В таблице маршрутизации должны присутствовать оба nexthop. Пример ожидаемой
формы маршрута:

```text
default proto bgp metric 20
        nexthop via 10.255.1.2 dev wg-exit1 weight 1
        nexthop via 10.255.2.2 dev wg-exit2 weight 1
```

Команды `ip route get` с разными source port имитируют разные TCP-соединения.
При достаточном количестве разных портов в результате должны встречаться оба
`wg-exitN`. Два конкретных порта могут случайно попасть на один nexthop.

Для проверки динамического исключения шлюза:

```bash
sudo systemctl stop wg-quick@wg-exit1
sudo vtysh -c 'show bgp ipv4 unicast 0.0.0.0/0'
ip route show default
```

После завершения BGP/BFD convergence в default route должен остаться только
`10.255.2.2 dev wg-exit2`. После повторного запуска `wg-exit1` оба nexthop должны
появиться снова.

## Изменение состава шлюзов на ходу

Добавление external-3 состоит из четырех операций:

1. Создать отдельный интерфейс `wg-exit3` и отдельную `/30`-сеть.
2. Добавить BGP neighbor на server и external-3.
3. Разрешить external-3 анонсировать только `0.0.0.0/0`.
4. Запустить WireGuard и дождаться установления BGP-сессии.

После получения равноправного default route FRR автоматически добавит новый
nexthop в ECMP. Останавливать существующие туннели или очищать маршруты не нужно.
Для планового вывода шлюза сначала прекращают анонс default route, дожидаются
исключения nexthop и только затем останавливают WireGuard.

WireGuard peer и BGP neighbor все равно должны быть зарегистрированы заранее.
BGP динамически управляет доступностью уже зарегистрированных шлюзов, но сам не
создает WireGuard peers и не передает ключи.

## Устойчивое перераспределение потоков

Обычный Linux ECMP сохраняет один путь для пакетов с одинаковым 5-tuple, но при
изменении количества nexthop хеш может назначить часть существующих соединений
другому шлюзу. Из-за смены публичного NAT-адреса такие соединения оборвутся.

На Linux 5.19 и новее рекомендуется использовать resilient nexthop group. В
современных версиях FRR это включается глобально в `/etc/frr/frr.conf`:

```text
zebra nexthop-group resilience buckets 256 idle-timer 60 unbalanced-timer 300
```

Resilient hashing хранит назначение bucket -> nexthop. При удалении шлюза
переназначаются прежде всего его buckets, а при добавлении шлюза переносится
ограниченная часть buckets. Это уменьшает число оборванных соединений, но не
может сохранить соединения через отказавший шлюз: их NAT-состояние находилось на
этом шлюзе.

Команда `zebra nexthop-group resilience` отсутствует в FRR 8.4 из штатного
репозитория Debian 12. Для нее нужна более новая версия FRR. Стандартное ядро
Debian 12 версии 6.1 требование Linux 5.19 выполняет. До обновления FRR схема
работает как обычный динамический ECMP, но с большим риском переназначения
существующих потоков при изменении состава группы.

Проверка resilient-группы:

```bash
ip -d nexthop list
ip -s nexthop bucket show
```

## Балансировка по фактической загрузке

Равные ECMP-веса распределяют потоки, а не байты. Один большой поток способен
занять весь канал, даже если количество соединений на шлюзах одинаково. Для
шлюзов разной емкости сначала задаются статические веса пропорционально доступной
полосе, например `1:2` для каналов 100 и 200 Мбит/с.

Для автоматической балансировки по текущей загрузке нужен контроллер, который:

1. Считывает скорость, потери и задержку каждого `wg-exitN`.
2. Исключает шлюз после нескольких неудачных health-check, не ожидая загрузки.
3. Рассчитывает веса по свободной полосе с EWMA и гистерезисом.
4. Обновляет веса не чаще одного раза в 30-60 секунд.
5. Применяет новое распределение только к новым соединениям.

BGP и BFD являются источником состояния доступности, но сами не измеряют
загрузку интернет-канала. Частое изменение ECMP-весов без гистерезиса будет
переназначать потоки и ухудшит работу NAT. Если требуется строгая привязка
существующего соединения к шлюзу при любом изменении весов, на server нужен
`conntrack mark` и отдельная policy-routing table для каждого `wg-exitN`; это
должен настраивать и атомарно обновлять контроллер.

## Ограничение BFD

BFD быстро обнаруживает отказ узла, WireGuard-туннеля или пути между BGP
neighbors. Он не гарантирует, что внешний узел действительно имеет доступ ко
всему интернету. Если endpoint server доступен, но upstream external-1 не
работает, BGP-сессия может остаться поднятой.

Для такого случая нужен отдельный health-check через `eth0`. Он должен удалять
default route или отзывать его BGP-анонс после нескольких неудачных проверок и
возвращать после устойчивого восстановления связи.

## Ссылки

- [WireGuard AllowedIPs](https://git.zx2c4.com/wireguard-tools/about/src/man/wg.8)
- [wg-quick и параметр Table](https://git.zx2c4.com/wireguard-tools/about/src/man/wg-quick.8)
- [FRR BGP](https://docs.frrouting.org/en/latest/bgp.html)
- [FRR BFD](https://docs.frrouting.org/en/latest/bfd.html)
- [FRR nexthop groups](https://docs.frrouting.org/en/latest/nexthop_groups.html)
- [FRR Zebra](https://docs.frrouting.org/en/latest/zebra.html)
- [Linux resilient nexthop groups](https://docs.kernel.org/networking/nexthop-group-resilient.html)
- [Linux IP sysctl](https://docs.kernel.org/networking/ip-sysctl.html)
