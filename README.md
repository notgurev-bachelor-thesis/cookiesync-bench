# Cookiesync benchmark

Модуль, нагружающий producer данными.

## Использование

Создать нагрузку: 

```shell
wrk -t 16 -c 30 -d 60m -s ./prof/cs.lua "http://cl-hot1-1.moevideo.net:8080"
```
