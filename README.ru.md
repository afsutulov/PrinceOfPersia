# Prince of Persia

<img src="icon.png" width="96" height="96" alt="Иконка Prince of Persia">

[English](README.md)

Prince of Persia — настольная версия классического платформера **Prince of Persia** под Windows, Linux, macOS. Исследуйте подземелье и дворец, перепрыгивайте пропасти, забирайтесь на уступы, избегайте шипов и закрывающихся ворот, собирайте зелья, сражайтесь со стражниками и спасите принцессу.

## Скриншоты

Настоящие снимки работающей Go-версии игры.

**Подземелье**

![Игра Prince в подземелье](screenshots/dungeon.png)

**Дворец**

![Игра Prince во дворце](screenshots/palace.png)

## Требования

- Go **1.25 или новее**. Для первой сборки нужен доступ к сети для загрузки Go-модулей; при генерации иконки Windows EXE также загружается инструмент ресурсов.
- Windows x64; Linux x64/ARM64 с glibc; macOS 12 или новее на Intel/Apple Silicon.
- Рабочее графическое окружение и аудиоустройство. В Linux нужны X11 или XWayland, драйвер OpenGL и системные библиотеки ALSA/PulseAudio.

## Сборка и запуск

Выполняйте команды из корня проекта, где находятся `main.go` и `go.mod`.

### Windows x64 — EXE с иконкой

```powershell
$env:CGO_ENABLED = "0"
go generate .
go build -buildvcs=false -trimpath -ldflags="-s -w -H=windowsgui" -o Prince.exe .
.\Prince.exe
```

`go generate .` преобразует `icon.ico` в `icon_windows_amd64.syso`. Go автоматически включает этот ресурс в EXE, и файл получает иконку в Проводнике. `-H=windowsgui` отключает открытие консольного окна. Созданный `.syso` — результат сборки, его нет в архиве исходников. После замены `icon.ico` выполните генерацию повторно. После сборки для запуска достаточно только `Prince.exe`.

Используется инструмент [rsrc](https://github.com/akavel/rsrc), версия v0.10.2 зафиксирована в директиве `go:generate`. Одного флага `go build` для иконки файла Windows недостаточно: сначала нужно сгенерировать ресурс.

### Linux

```sh
CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w" -o Prince .
./Prince
```

PNG-иконка встраивается автоматически и используется окном игры и, при поддержке окружением рабочего стола, панелью задач. У ELF-файлов нет стандартного ресурса иконки для файлового менеджера. При установке ярлыка можно указать `icon.png` в файле `.desktop`; это необязательная интеграция с рабочим столом, для запуска игры она не требуется.

### macOS — отдельный исполняемый файл

```sh
CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w" -o Prince .
./Prince
```

Эта команда собирает игру для архитектуры текущего Mac. Отдельного исполняемого файла достаточно для игры. Для иконки в Finder/Dock и запуска двойным щелчком можно создать необязательный пакет `.app` с `icon.icns`:

```sh
mkdir -p Prince.app/Contents/MacOS Prince.app/Contents/Resources
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o Prince.app/Contents/MacOS/Prince .
cp icon.icns Prince.app/Contents/Resources/icon.icns
cat > Prince.app/Contents/Info.plist <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>CFBundleName</key><string>Prince</string>
  <key>CFBundleDisplayName</key><string>Prince</string>
  <key>CFBundleIdentifier</key><string>org.Prince.game</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleExecutable</key><string>Prince</string>
  <key>CFBundleIconFile</key><string>icon.icns</string>
  <key>CFBundleVersion</key><string>1.1.0</string>
  <key>CFBundleShortVersionString</key><string>1.1.0</string>
  <key>NSHighResolutionCapable</key><true/>
  <key>LSMinimumSystemVersion</key><string>12.0</string>
</dict></plist>
PLIST
open Prince.app
```

`.app` — пакет macOS с исполняемым файлом, иконкой и метаданными. Это отдельный вариант оформления, дополнительно к поставке одним файлом. Команды создают локальное неподписанное приложение; для публичного распространения с подписью Developer ID и нотарификацией используются инструменты Apple.

### Кросс-компиляция

Из POSIX-оболочки в Linux или macOS:

```sh
# Windows x64
# Выполните генератор на исходной ОС до задания GOOS/GOARCH.
go generate .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -H=windowsgui" -o Prince-windows-amd64.exe .

# Linux x64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o Prince-linux-amd64 .

# Linux ARM64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o Prince-linux-arm64 .

# macOS Apple Silicon
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o Prince-macos-arm64 .

# macOS Intel
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o Prince-macos-amd64 .
```

Выбирайте файл нужной системы и архитектуры. Во всех случаях cgo отключён. Имя `.syso` ограничивает этот ресурс Windows amd64; при сборке Linux и macOS он игнорируется. В PowerShell задавайте `GOOS`, `GOARCH` и `CGO_ENABLED` через переменные `$env:` перед `go build`; для `go generate .` должна быть выбрана исходная ОС, чтобы генератор ресурсов мог запуститься.

## Управление

| Клавиша | Действие |
| --- | --- |
| Влево / Вправо | Движение; разворот из положения стоя |
| Вверх | Прыжок; подъём или подтягивание в подходящем месте; блокирование удара в бою |
| Вниз | Приседание; спуск или отпускание уступа; убрать меч в бою |
| Shift + направление | Осторожный шаг; захват уступа при прыжке или падении |
| Shift | Подобрать меч, выпить зелье или выполнить действие рядом; ударить в бою |
| Escape | Пауза / продолжение |
| Ctrl+A | Перезапустить текущий уровень |
| Ctrl+R | Вернуться к началу игры |
| Ctrl+G | Сохранить игру на уровнях 1–13 |
| Ctrl+L | Загрузить сохранение с начального экрана |
| Ctrl+S | Включить / выключить звук |
| Ctrl+Q | Выйти |
| Alt+Enter | Переключить полноэкранный режим |

Чтобы получить меч на первом уровне, встаньте рядом с ним и удерживайте **Shift**. Если стоите далеко, сделайте осторожный шаг к мечу. Для игры также можно использовать поддерживаемые геймпады. В сочетаниях клавиш используется **Control**, в том числе на macOS.

Обычное сохранение содержит номер уровня, оставшееся время и здоровье в начале уровня. При загрузке уровень начинается заново; произвольная позиция внутри комнаты не восстанавливается.

## Сохранения и игровые данные

В исполняемом файле находится всё необходимое для старта. При первом запуске создаётся пользовательский профиль для сохранений, рекордов и рабочих копий встроенных ресурсов. Рядом с исполняемым файлом дополнительные файлы не нужны.

| Система | Профиль по умолчанию |
| --- | --- |
| Windows | `%APPDATA%\Prince` |
| Linux | `$XDG_CONFIG_HOME/Prince` или `~/.config/Prince` |
| macOS | `~/Library/Application Support/Prince` |

Сохранения находятся в `dos14/PRINCE.SAV`, рекорды — в `dos14/PRINCE.HOF`. Другой профиль задаётся так:

```sh
./Prince -profile /путь/к/профилю
```

В Windows:

```powershell
.\Prince.exe -profile C:\Games\PrinceProfile
```

Используйте короткий путь к профилю, не более 180 байт UTF-8. Раздельные профили сохраняют независимые игры. Рабочие копии встроенных ресурсов восстанавливаются, если они отсутствуют или повреждены.

## Лицензия и авторы

Исходный код движка: **GPL-3.0-or-later**, см. [LICENSE](LICENSE). Авторы движка и зависимостей, а также их лицензии указаны в [licenses](licenses), включая [NOTICE.md](licenses/NOTICE.md). Оригинальные графика, уровни и звук Prince of Persia сохраняют исходные права; лицензия кода не предоставляет отдельную лицензию на эти ресурсы. Prince of Persia создана Джорданом Мекнером и выпущена Brøderbund.
