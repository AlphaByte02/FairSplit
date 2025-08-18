# FairSplit

FairSplit is a modern web application that helps you **track group expenses** and fairly split payments among friends. Each user can create sessions, record transactions, invite other participants, and finally calculate intermediate or final balances with minimized money transfers.

The project stack:

-   **Go** (Fiber v3)
-   **PostgreSQL** (with UUIDv7)
-   **sqlc** (type-safe queries)
-   **templ** (templating engine)
-   **UnoCSS** (with Tailwind-compatible reset)
-   **Alpine.js** & **HTMX** (lightweight interactivity)
-   **Air** (live reload during development)

---

## 🚀 Features

-   Simple user login (username-based) (for now™)
-   Create and join sessions
-   Add transactions (including who paid for whom)
-   View intermediate balances per user
-   View minimized final settlements
-   Lock sessions when finished
-   Responsive dark-mode UI with modern "glass" aesthetics

---

## 🛠️ Development

### Requirements

-   Go 1.22+
-   Docker & Docker Compose

### Setup

1. Clone the repository:

    ```bash
    git clone https://github.com/AlphaByte02/FairSplit.git
    cd FairSplit
    ```

2. Start services with Docker Compose:

    ```bash
    docker compose up -d --build
    ```

    This starts:

    - `app` on port 8080 (3000 for `air` with proxy)
    - `postgres` (Postgres)
    - `adminer` on port 8081 (Adminer)

3. Apply the migrations in the [migration folder](sql/migrations)

4. Access the app at [http://localhost:3000](http://localhost:3000).

5. During development, changes are hot-reloaded automatically via **Air**.

### Useful Commands/Tips

-   If you run `air` on your machine (not docker) you can set the environment variables by copying the `.env.example` file to `.env` and compile that file as you like

-   If you change any .sql files **run sqlc** to recompile the queries

    ```bash
    sqlc generate
    ```

-   You can recompile manualy the templ files with:
    ```bash
    go tool templ generate
    ```
    > If you use air there is no need to do it yourself (unless you made a small change when the service is down)

---

## 📦 Deployment

1. Set the appropiate environment variables in [docker-compose.prod.yml](docker-compose.prod.yml). Remember to set the postgres variable with secure value

2. Start the Postgres server

    ```bash
    docker compose -f docker-compose.prod.yml up -d postgres
    ```

3. Apply the migrations in the [migration folder](sql/migrations)

4. Start services in detached mode (use `--build` only the first time):

    ```bash
    docker compose -f docker-compose.prod.yml up -d --build
    ```

By default, the app runs on port **8080**.

---

## 🤝 Contributing

This project is for me a way to learn more about variuos way to do web dev.

If you have any suggestions for improvements, new features, or bug fixes open a Issue or Pull Requests are always welcome:

1. Fork the repo
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request 🎉

Please make sure to update documentation and follow code style conventions.

---

## 📜 License

This project is licensed under the [GNU GPLv3 License](https://choosealicense.com/licenses/gpl-3.0/).

---

### ✨ Acknowledgements

-   [Fiber](https://gofiber.io/)
-   [templ](https://templ.guide/)
-   [UnoCSS](https://unocss.dev/)
-   [Alpine.js](https://alpinejs.dev/)
-   [HTMX](https://htmx.org/)
-   [sqlc](https://sqlc.dev/)
