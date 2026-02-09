# Expense Tracker

Expense Tracker is a Go-based web service designed to help users efficiently manage and monitor their expenses. The project follows a clean architecture with separate handler, service, and repository layers, ensuring maintainable and scalable code. It features JWT-based authentication for secure access and provides endpoints to create, read, update, and delete expense records.

The application can be run locally by cloning the repository, installing dependencies using `go mod download`, and starting the server with `go run ./cmd/api`. It listens on the port defined by the `PORT` environment variable (default: 8080), and requires `JWT_SECRET` to be set for authentication. Database credentials can be configured as needed.

For deployment, the project includes a Dockerfile that allows building and running the application in a container with `docker build -t expense-tracker .` and `docker run -p 8080:8080 expense-tracker`. The app is also fully compatible with Render for automated cloud deployment.

This project demonstrates skills in Go development, clean architecture, API design, authentication, Dockerization, and cloud deployment, making it a strong example for recruiters.

## License

MIT License
