<div align="center">

<img src="https://bytelyon-public.s3.amazonaws.com/bytelyon-logo-lite.png" alt="Logo">

![Static Badge](https://img.shields.io/badge/Laravel-13.7-red?logo=laravel&color=red)
![Static Badge](https://img.shields.io/badge/PHP-8.5-777BB4?logo=php)
[![Laravel Forge Site Deployment Status](https://img.shields.io/endpoint?url=https%3A%2F%2Fforge.laravel.com%2Fsite-badges%2F52d486b6-7617-4880-a09c-501c69ad25f0%3Fdate%3D1&style=flat)](https://forge.laravel.com/nelsw/merciful-night-3dc/3268752)
![Golang 1.27](https://img.shields.io/static/v1?message=1.27&logo=go&labelColor=grey&color=00ADD8&label=%20)
![Python 3.14](https://img.shields.io/static/v1?message=3.14&logo=python&labelColor=grey&color=blue&label=%20)
</div>

## About ByteLyon

---

```bash
bytelyon/
├── .github/
│   └── workflows/
│       ├── web-ci.yml
│       ├── mgr-ci.yml
│       └── bro-ci.yml
│
├── apps/
│   ├── web/                    # Laravel
│   │   ├── app/
│   │   ├── routes/
│   │   ├── resources/
│   │   ├── composer.json
│   │   └── artisan
│   │
│   ├── mgr/                    # Goravel
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   ├── pkg/                # exportable packages (if any)
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── bro/                    # FastAPI
│       ├── src/
│       │   └── bro/
│       ├── pyproject.toml
│       └── poetry.lock         # or requirements.txt / uv.lock
│
├── packages/                   # shared, cross-language contracts
│   └── proto/                  # gRPC/protobuf definitions (if using gRPC)
│       └── *.proto
│
├── infra/
│   ├── docker/
│   │   ├── web.Dockerfile
│   │   ├── mgr.Dockerfile
│   │   └── bro.Dockerfile
│   ├── compose.yml
│   ├── compose.prod.yml
│   └── compose.test.yml
│
└── README.md
```
