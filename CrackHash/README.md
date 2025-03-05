### Пример 1. Поиск простого слова «a»

**Отправка задачи (maxLength = 1):**

```cmd
curl -X POST -H "Content-Type: application/json" -d "{\"hash\":\"0cc175b9c0f1b6a831c399e269772661\", \"maxLength\":1}" http://localhost:8080/api/hash/crack
```

**Пример ожидаемого ответа от менеджера:**

```json
{"requestId":"<some-uuid>"}
```

**Проверка статуса (подождите пару секунд, затем выполните):**

```cmd
curl "http://localhost:8080/api/hash/status?requestId=<ВАШ_REQUEST_ID>"
```

**Ожидаемый результат:**

```json
{"status":"READY","data":["a"]}
```

---

### Пример 2. Поиск слова, которое не входит в пространство (ожидается пустой результат)

При этом задаём `maxLength = 3`.  
Поскольку «test» имеет длину 4, ни одна из сгенерированных комбинаций не совпадёт с данным хэшом.

**Отправка задачи:**

```cmd
curl -X POST -H "Content-Type: application/json" -d "{\"hash\":\"098f6bcd4621d373cade4e832627b4f6\", \"maxLength\":3}" http://localhost:8080/api/hash/crack
```

**Проверка статуса:**

```cmd
curl "http://localhost:8080/api/hash/status?requestId=<ВАШ_REQUEST_ID>"
```

**Ожидаемый результат (READY, но data будет пустым):**

```json
{"status":"READY","data":[]}
```

---

### Пример 3. Поиск слова «abc» с maxLength = 4

Хэш для строки «abc»:
```
900150983cd24fb0d6963f7d28e17f72
```
При maxLength = 4 система генерирует все комбинации от длины 1 до 4 (общее число около 1,7 млн), что может занять немного больше времени.

**Отправка задачи:**

```cmd
curl -X POST -H "Content-Type: application/json" -d "{\"hash\":\"900150983cd24fb0d6963f7d28e17f72\", \"maxLength\":4}" http://localhost:8080/api/hash/crack
```

**Проверка статуса (подождите 30 секунд или чуть больше):**

```cmd
curl "http://localhost:8080/api/hash/status?requestId=<ВАШ_REQUEST_ID>"
```

**Ожидаемый результат:**

Если всё прошло успешно, статус должен стать READY, а в data появится найденное слово:

```json
{"status":"READY","data":["abc"]}
```

---

### Пример 4. Поиск слова «abc» при ограничении maxLength = 2

В этом примере мы используем тот же хэш для «abc» (900150983cd24fb0d6963f7d28e17f72), но задаём максимальную длину 2. Поскольку «abc» имеет длину 3, система не сможет его сгенерировать.

**Отправка задачи:**

```cmd
curl -X POST -H "Content-Type: application/json" -d "{\"hash\":\"900150983cd24fb0d6963f7d28e17f72\", \"maxLength\":2}" http://localhost:8080/api/hash/crack
```

**Проверка статуса:**

```cmd
curl "http://localhost:8080/api/hash/status?requestId=<ВАШ_REQUEST_ID>"
```

**Ожидаемый результат (READY с пустым data):**

```json
{"status":"READY","data":[]}
```