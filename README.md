# Accumulative Loyalty System

## Overview

It is a project for Yandex Practicum course **Advanced Go developer**.

The project is divided into two parts: an accrual system and a core system.

The accrual system provides an opportunity to register information about the reward for the order and to register the order for the calculation of bonus points.

The core system allows the user to register and start registering their completed orders, then the core system accesses the accrual system to obtain the number of points to be credited to the user, as well as the user can write off the accumulated points for future orders.
## Technologies

- **Language:** Go

- **Database:** Postgres

- **Documentation:** Swagger 2.0

## Getting Started

To get started with the Accumulative Loyalty System, follow these steps:

1. **Clone the Repository:**
```shell
git clone https://github.com/MowlCoder/accumulative-loyalty-system.git
```

2. **Install Dependencies:**
```shell
go mod tidy
```

3. **Configure Settings:** Create an `.env` file and populate it based on the `.env.example` file

4. **Run application:**
```shell
go run ./cmd/gophermart/main.go
```
```shell
go run ./cmd/accrual/main.go
```

## Documentation

Documentation is available in the [docs](/docs) directory or at `/swagger/index.html` endpoint.

## Contact

If you have any questions or need assistance, please don't hesitate to reach out at **maikezseller@gmail.com**.
