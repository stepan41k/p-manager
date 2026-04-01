# Пример

### 1. Структура данных
```go
type Entry struct {
    Service  string `json:"service"`  // например, "github.com"
    Login    string `json:"login"`
    Password string `json:"password"`
    Note     string `json:"note"`
}

type Vault struct {
    Entries []Entry `json:"entries"`
}
```

### 2. Логика добавления пароля
- Запрос мастер-пароля: Пользователь вводит мастер-пароль в терминале.
- Download: Скачиваем существующий файл vault.enc из Selectel S3 в память (массив байт). Если файла еще нет — создаем пустой объект Vault.
- Decrypt: Превращаем байты из S3 в чистый JSON с помощью мастер-пароля.
- Десериализация: json.Unmarshal превращает JSON в структуру Vault.
- Модификация: Добавляем в массив Entries новую запись.
- Сериализация: json.Marshal превращает обновленный Vault обратно в JSON.
- Encrypt: Шифруем JSON новым солью и нонсом.
- Upload: Отправляем зашифрованные байты в Selectel S3, заменяя старый файл.

### 3. Логика получения пароля
- Поиск: Пользователь выбирает сервис, например: github.
- Запрос мастер-пароля: без него не расшифровать.
- Download: Получаем vault.enc из S3.
- Decrypt: Получаем чистый JSON.
- Поиск в памяти: Проходим циклом по Vault.Entries
- Очистка: Очень важно — после вывода пароля в терминал, затрите переменные в памяти (обнулять слайсы байт).

<img width="615" height="522" alt="image" src="https://github.com/user-attachments/assets/5baf85aa-3e6a-4db7-8835-691da8c3e768" />
