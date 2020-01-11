# marslang
интерпретатор для будущей игры

Попытка создать интерпретаторя для простого, но строгого языка Marslang, имеющим следующие особенности:
* 1 стейтмент на одну строку. Стейтмент это выражение, которое не возвращает результат
* исходя из пункта выше, стейтменты не нужно завершать символом `;`
* язык со строгой типизацией, но без объявления переменных - тип определяется при инициализации, и не может быть впоследствии изменен
* нельзя проводить операции над разными типами, даже если это float и int - будет ошибка. нужно использовать приведение типов типа `a = 3 + int(4.5)`
* функции всегда задаются как переменные для простоты синтаксиса
* примеры простых программ:
```
sum = fn(int x, int y) int {
   return x + y
}
a = sum(2, 5)
```

# TODO
* Улучшить сообщения об ошибке - при выполнении - стэктрейс + номер строки/позиция
* control flow - if
* control flow - for
* Тип bool
* Тип string
* Тип array
* Поддержка структур
* Поддержка пакетов
* Расчет костов исполнения программы
* Контроль глубины стэка вызовов
* Бенчмарки - трэкинг производительности интерпретатора
* Импорты
