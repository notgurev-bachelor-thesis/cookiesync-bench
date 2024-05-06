# Cookiesync benchmark

Модуль, нагружающий producer данными.

## Использование

Создать нагрузку: 

```shell
wrk -t 16 -c 30 -d 60m -s ./prof/cs.lua "http://cl-hot1-1.moevideo.net:8080"
```

## Другие способы

```shell
./bin/bombardier -c 2048 -d 5s -H 'Cookie: uid=12345678901234567890' 'http://cl-hot1-1.moevideo.net:8080?b=10000&d=10'
```
