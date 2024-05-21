# Cookiesync benchmark

Модуль, нагружающий producer данными.

## Использование

Создать нагрузку:

```shell
wrk -t 16 -c 30 -d 60m -s ./prof/cs.lua "http://cl-hot1-1.moevideo.net:8080"
```

## Другие способы

Bombardier:

```shell
./bombardier -c 512 -d 5s -H 'Cookie: uid=12345678901234567890' 'http://cl-hot1-1.moevideo.net:8080?b=10000&d=10'
```

Vegeta:

```shell
echo "GET http://cl-hot1-1.moevideo.net:8080?b=10000&d=10" | ./vegeta attack -duration=10s -rate=0 -connections=2000 -max-connections=2000 -keepalive -max-workers=10240 -workers=1024 | tee results.bin | ./vegeta report
```
