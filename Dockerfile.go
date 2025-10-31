# ETAPA 1: BUILDER (Compilación)
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compila el ejecutable estático
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./main.go

# ----------------------------------------
# ETAPA 2: PRODUCTION (Contenedor Final Mínimo)
FROM alpine:latest

WORKDIR /root/

# Copia el binario compilado
COPY --from=builder /app/app .

EXPOSE 3000

# Comando para iniciar la aplicación
CMD ["./app"]