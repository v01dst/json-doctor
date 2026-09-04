<div align="center">

# 🩺 json-doctor

**Validate, inspect and fix JSON — from the terminal, instantly.**

[![CI](https://github.com/v01dst/json-doctor/actions/workflows/ci.yml/badge.svg)](https://github.com/v01dst/json-doctor/actions/workflows/ci.yml)
![License](https://img.shields.io/badge/license-MIT-8A2BE2)
![Go](https://img.shields.io/badge/go-1.27-00ADD8?logo=go&logoColor=white)
![Deps](https://img.shields.io/badge/dependencies-stdlib%20only-00c853)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

`validate` · `pretty / minify` · `stats` · `dot-path queries`

</div>

---

## ✨ Features

- **✅ Validation with location** — `--explain` shows the exact line and column of syntax errors
- **💅 Pretty / minify** — 2-space reformat or whitespace-strip in one pass
- **📊 Structural stats** — key counts, max depth, type histogram, entropy for your JSON
- **🔎 Dot-path queries** — `users.0.name`, array indexes included
- **🪶 Pure stdlib** — zero dependencies, single ~2 MB binary
- **🔁 Unix-friendly** — reads stdin or file, pipes anywhere

## 🚀 Quick Start

```bash
git clone https://github.com/v01dst/json-doctor
cd json-doctor
go build -o json-doctor .
```

Or with Docker:

```bash
docker build -t json-doctor .
docker run --rm -i json-doctor --pretty < data.json
```

## 📖 Usage

```
json-doctor [flags] [file.json]     (stdin if no file)

  --pretty     reformat with 2-space indent
  --minify     strip all whitespace
  --stats      structural statistics
  -q PATH      dot-path query (users.0.name)
  --explain    on error, print line/column
  --version    print version
```

### Examples

```bash
$ echo '{"name":"ada","langs":["go","py"]}' | json-doctor --pretty
{
  "name": "ada",
  "langs": [
    "go",
    "py"
  ]
}

$ echo '{"broken": tru}' | json-doctor --explain
error: invalid character '}' in literal true (expecting 'e') (line 1, column 16)

$ json-doctor -q users.0.email data.json
"ada@example.com"

$ json-doctor --stats big.json
{
  "keys": 3,
  "maxDepth": 4,
  "arrays": 1,
  ...
}
```

| Exit code | Meaning            |
|-----------|--------------------|
| `0`       | Success            |
| `1`       | Invalid JSON / bad query |

## 🧱 Tech Stack

| Layer     | Tech                  |
|-----------|-----------------------|
| Language  | Go 1.27 (stdlib only) |
| Testing   | go test               |
| Packaging | Docker (multi-stage)  |
| CI        | GitHub Actions        |

---

<div align="center">

Built with ⚡ by **v01dst**

[![GitHub](https://img.shields.io/badge/github-v01dst-181717?logo=github)](https://github.com/v01dst)
[![Discord](https://img.shields.io/badge/discord-9p.1-5865F2?logo=discord&logoColor=white)](https://discord.com/users/9p.1)

*Project 009 / 99 — The Loop*

</div>
