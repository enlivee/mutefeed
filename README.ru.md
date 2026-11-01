# MuteFeed

<p align="center">
  <a href="README.md">English</a> · <b>Русский</b>
</p>

```text
  __  __       _       ______            _ 
 |  \/  |     | |     |  ____|          | |
 | \  / |_   _| |_ ___| |__ ___  ___  __| |
 | |\/| | | | | __/ _ \  __/ _ \/ _ \/ _` |
 | |  | | |_| | ||  __/ | |  __/  __/ (_| |
 |_|  |_|\__,_|\__\___|_|  \___|\___|\__,_|
```

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![CI Status](https://img.shields.io/badge/CI-Passing-brightgreen?style=flat-square&logo=github-actions&logoColor=white)](https://github.com/enlivee/mutefeed/actions)
[![Docker Alpine](https://img.shields.io/badge/Docker%20Alpine-%3C%2015MB-2496ED?logo=docker&logoColor=white&style=flat-square)](https://github.com/enlivee/mutefeed/pkgs/container/mutefeed)
[![Dependencies](https://img.shields.io/badge/Dependencies-Zero%20External-success?style=flat-square)](#архитектура-и-философия)
[![Regex Safety](https://img.shields.io/badge/Regex-RE2%20Unicode%20Safe-blueviolet?style=flat-square)](#пресеты-и-эвристики)

**Активное шумоподавление для вашей информационной диеты.**

> Проект родился на основе **[самой востребованной идеи за 16 лет истории Hacker News](https://stillunbuilt.shiptiffin.app/i/a-mute-button-for-news-feeds/)** (45 независимых запросов с 2008 по 2024 год).
>
> MuteFeed — это сверхбыстрый фильтрующий прокси для RSS/Atom лент в виде одного бинарника. Он вырезает нейросетевой мусор (AI slop), обрубки платных статей (пейволлы), кричащий кликбейты и трекеры слежки до того, как они попадут в вашу читалку.

### 🌐 Онлайн-сервис (Без установки и настройки)
**👉 [mutefeed.onrender.com](https://mutefeed.onrender.com)**  
*Вставьте ссылку на любой RSS-канал, выберите, что заглушить, и получите чистый поток для NetNewsWire, Feedly или Reeder за 5 секунд. 100% бесплатно, без регистрации и деплоя.*

---

## Оглавление

- [Онлайн-сервис (Без установки)](#-онлайн-сервис-без-установки-и-настройки)
- [Подписка в ридерах в 1 клик](#подписка-в-ридерах-в-1-клик)
- [Чистый IT-дайджест (Super-Feed)](#чистый-it-дайджест-super-feed)
- [Свои ключевые слова для заглушки](#свои-ключевые-слова-для-заглушки)
- [Динамические SVG-бейджи чистоты](#динамические-svg-бейджи-чистоты)
- [Пакетная очистка OPML подписок](#пакетная-очистка-opml-подписок)
- [Архитектура и философия](#архитектура-и-философия)
- [Пресеты и эвристики](#пресеты-и-эвристики)
- [Тесты производительности (15 000+ QPS)](#тесты-производительности-15-000-qps)
- [Установка утилиты (curl / Homebrew)](#установка-утилиты-curl--homebrew)
- [☕ Поддержать автора](#-поддержать-автора)
- [Лицензия](#лицензия)

---

## Подписка в ридерах в 1 клик

MuteFeed отдает стандартный RSS 2.0 XML и работает из коробки с любыми читалками (NetNewsWire, Reeder, Feedly, Inoreader, Miniflux, Thunderbird).

| Ридер | Платформа | Протокол | Пример ссылки для вставки |
|---|---|---|---|
| **NetNewsWire** | macOS / iOS | `feed://` | `feed://mutefeed.onrender.com/feed?url=https%3A%2F%2Fnews.ycombinator.com%2Frss` |
| **Reeder** | macOS / iOS | `feed://` / `reeder://` | `feed://mutefeed.onrender.com/feed?url=https%3A%2F%2Fhabr.com%2Fru%2Frss%2Fall%2Fall%2F` |
| **Feedly** | Web / Mobile | Web URL | Добавьте сгенерированный `https://mutefeed.onrender.com/feed?...` |
| **Inoreader** | Web / Mobile | Web URL | Добавьте сгенерированный `https://mutefeed.onrender.com/feed?...` |
| **Miniflux / FreshRSS** | Self-hosted | HTTP URL | Вставьте чистый URL прокси прямо в веб-интерфейс |

---

## Чистый IT-дайджест (Super-Feed)

Если вы не хотите подписываться на десятки каналов по отдельности, MuteFeed предоставляет готовую единую супер-ленту:

```
https://mutefeed.onrender.com/feed?bundle=tech
```

Она на лету объединяет **Hacker News**, **TechCrunch**, **The Verge**, **Ars Technica** и **Lobste.rs**, дедуплицирует одинаковые новости по заголовкам и ссылкам и вырезает рекламу и кликбейт.

Для русскоязычных IT-новостей доступен бандл:
```
https://mutefeed.onrender.com/feed?bundle=ru-tech
```
(Хабр, OpenNet, 3DNews).

---

## Пакетная очистка OPML подписок

У вас уже есть 50+ подписок в читалке? Не нужно переносить их вручную:
1. Экспортируйте файл `.opml` из вашего ридера (NetNewsWire, Reeder, Feedly).
2. Откройте [mutefeed.onrender.com](https://mutefeed.onrender.com) и нажмите кнопку **«Пакетный OPML»**.
3. Перетащите ваш файл в браузер.
4. MuteFeed перезапишет все адреса на чистые фильтрующие ссылки и скачает готовый `cleaned_subscriptions.opml`.
5. Импортируйте его обратно в ридер — готово!

---

## Пресеты и эвристики

| Пресет | Что отсекает | Примеры источников и триггеров |
|---|---|---|
| `ai-slop` | Вода и клише от ChatGPT | *«в современном быстро меняющемся мире»*, *«delve into»*, *«testament to»*, *«game-changer»* |
| `paywalls` | Жесткие пейволлы и платные обрубки | Ведомости, Forbes, Medium, WSJ, Bloomberg, FT, Wired |
| `clickbait` | Кликбейты и капслок | *«ШОК! Вы не поверите...»*, *«Вот почему X умер»*, кричащие заголовки |
| `sponsored` | Нативная реклама и промо | Скрытые интеграции, партнерские ссылки, рекламные подборки |
| `crypto-hype` | Спекуляции и скам | Мемкоины, памп токенов, эирдропы, шиллинг |

### Свои ключевые слова для заглушки

Вы можете заглушить любую конкретную тему через параметр `&mute=`:
```
https://mutefeed.onrender.com/feed?url=https%3A%2F%2Fnews.ycombinator.com%2Frss&mute=крипта,выборы,gta6
```

---

## Архитектура и философия

1. **Никаких внешних баз данных**: MuteFeed — это фильтрующий прокси, а не читалка. База статей уже есть у вас в ридере. Прокси работает stateless и держит горячий кэш в RAM.
2. **Один бинарник на Go (15 МБ)**: потребляет менее 15 МБ оперативной памяти.
3. **Безопасные регулярные выражения (RE2 Unicode)**: корректная обработка границ слов на русском, английском и китайском языках без ложных срабатываний.
4. **Встроенный интерфейс**: современный дашборд на React 19 зашит прямо в Go-бинарник через `go:embed`.

---

## Тесты производительности (15 000+ QPS)

В стресс-тестах на 100 000 запросов сервис показал:
- **Пропускная способность**: `17 000+ запросов в секунду`
- **Медианная задержка (p50)**: `5.8 мс`
- **99-й перцентиль (p99)**: `13.6 мс`
- **Потребление памяти**: стабильные `~13 МБ Heap` без утечек

---

## Установка утилиты (curl / Homebrew)

Если вы хотите запустить локальный фильтр у себя на компьютере или сервере:

### 1. Установка в одну команду (macOS / Linux):
```bash
curl -fsSL https://mutefeed.onrender.com/install.sh | bash
```

### 2. Через Homebrew:
```bash
brew install enlivee/tap/mutefeed
brew services start mutefeed
```

### 3. Через Docker:
```bash
docker run -d -p 8899:8899 ghcr.io/enlivee/mutefeed:latest
```

---

## ☕ Поддержать автора

MuteFeed — полностью бесплатный проект с открытым исходным кодом. Если он бережёт ваше внимание, фильтрует мусор и экономит время, вы можете поддержать автора напрямую и 100% анонимно:

- 💎 **USDT / TON (The Open Network)**: `UQB7QlryTUSPKqlOaaolXI7iqn5PDAvXOWVKdGeEkHR7Ps-3` (через Telegram @CryptoBot или Tonkeeper)
- ⭐ **Звезда на GitHub**: поставьте звёздочку репозиторию [enlivee/mutefeed](https://github.com/enlivee/mutefeed) — это бесплатно и сильно помогает проекту!

---

## Лицензия

Проект распространяется под свободной лицензией [MIT](LICENSE).
