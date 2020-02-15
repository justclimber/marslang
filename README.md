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

пример программы для игры, базовые действия:
```
commands.move = 1.
commands.rotate = 0.5
commands.cannon.rotate = -0.8
commands.cannon.shoot = 0.1
```

пример программы для реальной игры:
```
ifempty obj = nearestByType(mech, objects, 3) {
   return 1
}
angleObj = angle(mech.x, mech.y, obj.x, obj.y)
angleMech = mech.angle
angleTo = angleObj - angleMech
if angleTo < -PI {
   angleTo = 2. * PI + angleTo
}
if angleTo > PI {
   angleTo = angleTo - 2. * PI
}

switch angleTo {
case > 1.:
   commands.rotate = 1.
case < -1.:
   commands.rotate = -1.
default:
   commands.rotate = angleTo
}

dist = distance(mech.x, mech.y, obj.x, obj.y)
if obj.type == 3 {
   commands.move = 1.
   return 1
}
if dist > 200. {
   commands.move = distance / 1000.
   if commands.move > 1. {
      commands.move = 1.
   }
}
toShoot = commands.rotate * commands.rotate * dist
if toShoot < 70. {
   commands.cannon.shoot = 0.1
   return 1
}
```

пример чуть посложнее:
```
ifempty xelon = getFirstTarget(1) {
   ifempty xelon = nearestByType(mech, objects, 3) {
      return 1
   }
   addTarget(xelon, 1)
}
angleXelon = angle(mech.x, mech.y, xelon.x, xelon.y)
angleMech = mech.angle
angleTo = angleXelon - angleMech
if angleTo < -PI {
   angleTo = 2. * PI + angleTo
}
if angleTo > PI {
   angleTo = angleTo - 2. * PI
}

switch angleTo {
case > 1.:
   commands.rotate = 1.
case < -1.:
   commands.rotate = -1.
default:
   commands.rotate = angleTo
}

commands.move = 1.

ifempty obj = getFirstTarget(2) {   
   ifempty obj = nearestByType(mech, objects, 2) {
      return 1
   }
   addTarget(obj, 2)
}

angleObj = angle(mech.x, mech.y, obj.x, obj.y)
angleMechWithCannon = mech.angle + mech.cAngle
if angleMechWithCannon > 2. * PI {
   angleMech = angleMechWithCannon - 2. * PI
}
cAngleTo = angleObj - angleMechWithCannon
if cAngleTo < -PI {
   cAngleTo = 2. * PI + cAngleTo
}
if cAngleTo > PI {
   cAngleTo = cAngleTo - 2. * PI
}

if cAngleTo * angleTo < 0. {
   cAngleTo = cAngleTo - angleTo
}

switch cAngleTo {
case > 1.:
   commands.cannon.rotate = 1.
case < -1.:
   commands.cannon.rotate = -1.
default:
   commands.cannon.rotate = cAngleTo
}

dist = distance(mech.x, mech.y, obj.x, obj.y)
toShoot = commands.cannon.rotate * commands.cannon.rotate * dist
if toShoot < 40. {
   commands.cannon.shoot = 0.1
   return 1
}

```

# TODO
* control flow - for
* Тип string
* Поддержка пакетов
* Контроль глубины стэка вызовов
* Бенчмарки - трэкинг производительности интерпретатора
* стэктрейс при ошибках
* Импорты
