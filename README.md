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
c = 10
if c > 8 {
    print(true)
} else {
    print(false)
}
struct point {
   float x
   float y
}
p = point{x = 1., y = 2.}
px = p.x
```

пример программы для игры:
```
obj = nearest(mech, objects)
angleObj = angle(mech.x, mech.y, obj.x, obj.y)
angleMech = mech.angle
angleTo = angleObj - angleMech
if angleTo < -PI {
   angleTo = 2. * PI + angleTo
}
if angleTo > PI {
   angleTo = angleTo - 2. * PI
}

switch {
case angleTo > 1.:
   mrThr = 1.
case angleTo < -1.:
   mrThr = -1.
default:
   mrThr = angleTo
}

distance = distance(mech.x, mech.y, obj.x, obj.y)
if distance > 200. {
   mThr = distance / 1000.
   if mThr > 1. {
      mThr = 1.
   }
}

if mrThr * mrThr * distance < 70. {
   shoot = 0.1
   return 1
}
```

# TODO
* control flow - for
* Тип string
* Поддержка пакетов
* Расчет костов исполнения программы
* Контроль глубины стэка вызовов
* Бенчмарки - трэкинг производительности интерпретатора
* стэктрейс при ошибках
* Импорты
