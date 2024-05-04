# financial-system

## Introduction

This repository contains a financial system application built with Docker, PostgreSQL, and Go. The application allows users to manage accounts and transactions.

## Launching the Application

To launch the application, follow these steps:

1. Navigate to the root directory of the project.
2. Run the following command to launch the application:

```bash
docker-compose up --build
```


## Execute the command: 
```bash
docker exec -it financial-system_db_1 psql -U postgres
```


## Adding the tables: 

```bash
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    balance DECIMAL(10, 2) NOT NULL
);
```

```bash
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    value DECIMAL(10, 2) NOT NULL,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    group_type VARCHAR(50) NOT NULL,
    account2_id INTEGER REFERENCES accounts(id),
    transaction_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Go to the host and use the service

localhost:8080

Enjoy!